package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"agent-hub/internal/config"
	contentHandler "agent-hub/internal/content/handler"
	contentRepo "agent-hub/internal/content/repository"
	contentService "agent-hub/internal/content/service"
	interactionHandler "agent-hub/internal/interaction/handler"
	interactionRepo "agent-hub/internal/interaction/repository"
	interactionService "agent-hub/internal/interaction/service"
	"agent-hub/internal/middleware"
	"agent-hub/internal/model"
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

type MySQLTestApp struct {
	Router      *gin.Engine
	DB          *gorm.DB
	DBName      string
	JWTSecret   []byte
	ExpireHours int
}

func NewMySQLTestApp(t *testing.T) *MySQLTestApp {
	t.Helper()

	cfg, err := loadConfigForTests()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.MySQL.Host == "" {
		t.Skip("mysql config missing")
	}

	dbName := fmt.Sprintf("agent_hub_test_%d", time.Now().UnixNano())
	adminDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
		cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Charset)

	admin, err := sql.Open("mysql", adminDSN)
	if err != nil {
		t.Skipf("mysql not available: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := admin.PingContext(ctx); err != nil {
		_ = admin.Close()
		t.Skipf("mysql not available: %v", err)
	}

	_, err = admin.ExecContext(ctx, "CREATE DATABASE `"+dbName+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	if err != nil {
		_ = admin.Close()
		t.Fatalf("create test database: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = admin.ExecContext(ctx, "DROP DATABASE `"+dbName+"`")
		_ = admin.Close()
	})

	testCfg := cfg.MySQL
	testCfg.Database = dbName
	db, err := gorm.Open(mysql.Open(testCfg.DSN()), &gorm.Config{})
	if err != nil {
		t.Fatalf("mysql open: %v", err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := model.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	seedDefaultCommunity(db)

	jwtSecret := []byte(cfg.JWT.Secret)
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("test-secret")
	}
	expireHours := cfg.JWT.ExpireHours
	if expireHours <= 0 {
		expireHours = 168
	}

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

	// Services + Handlers
	userSvc := userService.NewUserService(userRepository, agentRepository, userPointsRepo)
	authHandler := userHandler.NewAuthHandler(userSvc, jwtSecret, expireHours)
	agentHandler := userHandler.NewAgentHandler(userSvc, jwtSecret, expireHours)

	pointsSvc := pointsService.NewPointsService(pointsRepository)
	notificationSvc := notificationService.NewNotificationService(notificationRepository)

	contentSvc := contentService.NewContentService(postRepository, commentRepository, communityRepository, pointsSvc, notificationSvc)
	postHandler := contentHandler.NewPostHandler(contentSvc)
	commentHandler := contentHandler.NewCommentHandler(contentSvc)

	interactionSvc := interactionService.NewInteractionService(
		voteRepository, followRepository,
		postRepository, commentRepository,
		agentRepository,
		pointsSvc,
		notificationSvc,
	)
	voteHandler := interactionHandler.NewVoteHandler(interactionSvc)
	followHandler := interactionHandler.NewFollowHandler(interactionSvc)

	rankingSvc := rankingService.NewRankingService(rankingRepository)
	leaderboardHandler := rankingHandler.NewLeaderboardHandler(rankingSvc)

	searchRepository := searchRepo.NewSearchRepository(db)
	searchSvc := searchService.NewSearchService(searchRepository)
	sHandler := searchHandler.NewSearchHandler(searchSvc)

	notifHandler := notificationHandler.NewNotificationHandler(notificationSvc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/oauth/twitter", authHandler.OAuthTwitter)
			auth.GET("/oauth/twitter/callback", authHandler.OAuthTwitterCallback)
		}

		v1.POST("/agents", middleware.JWT(jwtSecret), agentHandler.Create)
		v1.GET("/agents/:agent_name", agentHandler.GetByName)
		v1.PUT("/me/agent", middleware.JWT(jwtSecret), agentHandler.UpdateMe)

		v1.POST("/posts", middleware.JWT(jwtSecret), postHandler.Create)
		v1.GET("/posts", postHandler.List)
		v1.GET("/posts/:post_id", postHandler.Get)
		v1.PUT("/posts/:post_id", middleware.JWT(jwtSecret), postHandler.Update)
		v1.DELETE("/posts/:post_id", middleware.JWT(jwtSecret), postHandler.Delete)

		v1.POST("/posts/:post_id/comments", middleware.JWT(jwtSecret), commentHandler.Create)
		v1.GET("/posts/:post_id/comments", commentHandler.List)
		v1.DELETE("/comments/:comment_id", middleware.JWT(jwtSecret), commentHandler.Delete)

		v1.POST("/posts/:post_id/vote", middleware.JWT(jwtSecret), voteHandler.PostVote)
		v1.POST("/comments/:comment_id/vote", middleware.JWT(jwtSecret), voteHandler.CommentVote)
		v1.POST("/agents/:agent_name/follow", middleware.JWT(jwtSecret), followHandler.Follow)

		v1.GET("/search", sHandler.Search)
		v1.GET("/leaderboard", leaderboardHandler.Get)

		v1.GET("/notifications", middleware.JWT(jwtSecret), notifHandler.List)
		v1.PATCH("/notifications/:id/read", middleware.JWT(jwtSecret), notifHandler.MarkRead)
		v1.POST("/notifications/read-all", middleware.JWT(jwtSecret), notifHandler.MarkAllRead)
	}

	return &MySQLTestApp{
		Router:      r,
		DB:          db,
		DBName:      dbName,
		JWTSecret:   jwtSecret,
		ExpireHours: expireHours,
	}
}

func seedDefaultCommunity(db *gorm.DB) {
	var count int64
	db.Model(&model.Community{}).Count(&count)
	if count > 0 {
		return
	}
	desc := "默认社区，欢迎讨论"
	db.Create(&model.Community{Name: "General", Description: &desc})
}

func loadConfigForTests() (*config.Config, error) {
	// 先尝试常规加载（当 go test 在 repo root 运行时可用）
	cfg, err := config.Load()
	if err == nil && cfg != nil && cfg.MySQL.Host != "" {
		return cfg, nil
	}

	// go test 通常以包目录为工作目录；这里用当前文件位置定位 repo root
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		// 退化到常规加载结果（可能为空）
		return cfg, err
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	configPath := filepath.Join(root, "configs", "config.yaml")

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 绑定环境变量（与 config.Load 保持一致）
	_ = v.BindEnv("server.port", "SERVER_PORT")
	_ = v.BindEnv("mysql.host", "MYSQL_HOST")
	_ = v.BindEnv("mysql.port", "MYSQL_PORT")
	_ = v.BindEnv("mysql.user", "MYSQL_USER")
	_ = v.BindEnv("mysql.password", "MYSQL_PASSWORD")
	_ = v.BindEnv("mysql.database", "MYSQL_DATABASE")
	_ = v.BindEnv("redis.addr", "REDIS_ADDR")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("jwt.secret", "JWT_SECRET")
	_ = v.BindEnv("jwt.expire_hours", "JWT_EXPIRE_HOURS")
	_ = v.BindEnv("log.level", "LOG_LEVEL")
	_ = v.BindEnv("log.format", "LOG_FORMAT")

	// 如果 repo root 有配置文件就读它；没有也不当错误（全靠 env）
	_ = v.ReadInConfig()

	var out config.Config
	if err2 := v.Unmarshal(&out); err2 != nil {
		return nil, err2
	}
	return &out, nil
}

