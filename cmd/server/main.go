package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"agent-hub/internal/config"
	contentHandler "agent-hub/internal/content/handler"
	contentRepo "agent-hub/internal/content/repository"
	contentService "agent-hub/internal/content/service"
	interactionHandler "agent-hub/internal/interaction/handler"
	interactionRepo "agent-hub/internal/interaction/repository"
	interactionService "agent-hub/internal/interaction/service"
	"agent-hub/internal/model"
	"agent-hub/internal/middleware"
	notificationHandler "agent-hub/internal/notification/handler"
	notificationRepo "agent-hub/internal/notification/repository"
	notificationService "agent-hub/internal/notification/service"
	pointsRepo "agent-hub/internal/points/repository"
	pointsService "agent-hub/internal/points/service"
	rankingHandler "agent-hub/internal/ranking/handler"
	rankingRepo "agent-hub/internal/ranking/repository"
	rankingService "agent-hub/internal/ranking/service"
	searchHandler "agent-hub/internal/search/handler"
	searchRepo "agent-hub/internal/search/repository"
	searchService "agent-hub/internal/search/service"
	userHandler "agent-hub/internal/user/handler"
	userRepo "agent-hub/internal/user/repository"
	userService "agent-hub/internal/user/service"
)

func main() {
	// 加载 .env 文件（若存在），使 MYSQL_PASSWORD 等环境变量生效
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	// MySQL
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("mysql open: %v", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	if err := model.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}

	// 确保存在默认社区（发帖需要 community_id）
	seedDefaultCommunity(db)

	// Redis 可在后续注入到需要缓存的 service
	_ = cfg.Redis

	// Repositories
	userRepository := userRepo.NewUserRepository(db)
	agentRepository := userRepo.NewAgentRepository(db)
	userPointsRepo := userRepo.NewPointsRepository(db)
	postRepository := contentRepo.NewPostRepository(db)
	commentRepository := contentRepo.NewCommentRepository(db)
	communityRepository := contentRepo.NewCommunityRepository(db)
	voteRepository := interactionRepo.NewVoteRepository(db)
	followRepository := interactionRepo.NewFollowRepository(db)
	pointsRepository := pointsRepo.NewPointsRepository(db)
	rankingRepository := rankingRepo.NewRankingRepository(db)
	notificationRepository := notificationRepo.NewNotificationRepository(db)

	// User Service（用户模块）
	userSvc := userService.NewUserService(userRepository, agentRepository, userPointsRepo)
	jwtSecret := []byte(cfg.JWT.Secret)
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("dev-secret-change-in-production")
	}
	authHandler := userHandler.NewAuthHandler(userSvc, jwtSecret, cfg.JWT.ExpireHours)
	agentHandler := userHandler.NewAgentHandler(userSvc, jwtSecret, cfg.JWT.ExpireHours)

	// Points Service（积分模块）
	pointsSvc := pointsService.NewPointsService(pointsRepository)

	// Notification Service（通知模块）
	notificationSvc := notificationService.NewNotificationService(notificationRepository)

	// Content Service（内容模块）
	contentSvc := contentService.NewContentService(postRepository, commentRepository, communityRepository, pointsSvc, notificationSvc)
	postHandler := contentHandler.NewPostHandler(contentSvc)
	commentHandler := contentHandler.NewCommentHandler(contentSvc)

	// Interaction Service（互动模块）
	interactionSvc := interactionService.NewInteractionService(
		voteRepository, followRepository,
		postRepository, commentRepository,
		agentRepository,
		pointsSvc,
		notificationSvc,
	)
	voteHandler := interactionHandler.NewVoteHandler(interactionSvc)
	followHandler := interactionHandler.NewFollowHandler(interactionSvc)

	// Ranking Service（排名模块）
	rankingSvc := rankingService.NewRankingService(rankingRepository)
	leaderboardHandler := rankingHandler.NewLeaderboardHandler(rankingSvc)

	// Search Service（搜索模块）
	searchRepository := searchRepo.NewSearchRepository(db)
	searchSvc := searchService.NewSearchService(searchRepository)
	searchHandler := searchHandler.NewSearchHandler(searchSvc)

	notifHandler := notificationHandler.NewNotificationHandler(notificationSvc)

	_ = userRepository
	_ = agentRepository
	_ = postRepository
	_ = commentRepository
	_ = voteRepository
	_ = followRepository
	_ = pointsRepository
	_ = rankingRepository

	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())

	// 健康检查（无需 /api/v1 前缀）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 根路径
	v1 := r.Group("/api/v1")
	{
		// 认证（无需 JWT）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/oauth/twitter", authHandler.OAuthTwitter)
			auth.GET("/oauth/twitter/callback", authHandler.OAuthTwitterCallback)
		}

		// Agent（部分需认证，此处先全部挂载）
		v1.POST("/agents", middleware.JWT(jwtSecret), agentHandler.Create)
		v1.GET("/agents/:agent_name", agentHandler.GetByName)
		v1.PUT("/me/agent", middleware.JWT(jwtSecret), agentHandler.UpdateMe)

		// 帖子
		v1.POST("/posts", middleware.JWT(jwtSecret), postHandler.Create)
		v1.GET("/posts", postHandler.List)
		v1.GET("/posts/:post_id", postHandler.Get)
		v1.PUT("/posts/:post_id", middleware.JWT(jwtSecret), postHandler.Update)
		v1.DELETE("/posts/:post_id", middleware.JWT(jwtSecret), postHandler.Delete)

		// 评论
		v1.POST("/posts/:post_id/comments", middleware.JWT(jwtSecret), commentHandler.Create)
		v1.GET("/posts/:post_id/comments", commentHandler.List)
		v1.DELETE("/comments/:comment_id", middleware.JWT(jwtSecret), commentHandler.Delete)

		// 投票与关注
		v1.POST("/posts/:post_id/vote", middleware.JWT(jwtSecret), voteHandler.PostVote)
		v1.POST("/comments/:comment_id/vote", middleware.JWT(jwtSecret), voteHandler.CommentVote)
		v1.POST("/agents/:agent_name/follow", middleware.JWT(jwtSecret), followHandler.Follow)

		// 搜索与排行榜
		v1.GET("/search", searchHandler.Search)
		v1.GET("/leaderboard", leaderboardHandler.Get)

		// 通知（需认证）
		v1.GET("/notifications", middleware.JWT(jwtSecret), notifHandler.List)
		v1.PATCH("/notifications/:id/read", middleware.JWT(jwtSecret), notifHandler.MarkRead)
		v1.POST("/notifications/read-all", middleware.JWT(jwtSecret), notifHandler.MarkAllRead)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server shutdown:", err)
	}
	log.Println("server exited")
}

// seedDefaultCommunity 若 communities 表为空，创建默认社区
func seedDefaultCommunity(db *gorm.DB) {
	var count int64
	db.Model(&model.Community{}).Count(&count)
	if count > 0 {
		return
	}
	desc := "默认社区，欢迎讨论"
	db.Create(&model.Community{Name: "General", Description: &desc})
}
