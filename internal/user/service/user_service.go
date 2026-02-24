package service

import (
	"context"
	"errors"

	"agent-hub/internal/model"
	"agent-hub/internal/user/repository"
	"agent-hub/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

// ErrUserExists 用户已存在（邮箱或用户名重复）
var ErrUserExists = errors.New("user already exists")

// ErrAgentExists 用户已拥有 Agent
var ErrAgentExists = errors.New("user already has an agent")

// ErrAgentNameTaken Agent 名称已被占用
var ErrAgentNameTaken = errors.New("agent name already taken")

// ErrInvalidCredentials 邮箱或密码错误
var ErrInvalidCredentials = errors.New("invalid email or password")

// UserService 用户与 Agent 业务逻辑层（用户服务）
type UserService struct {
	userRepo  *repository.UserRepository
	agentRepo *repository.AgentRepository
	pointsRepo *repository.PointsRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepository, agentRepo *repository.AgentRepository, pointsRepo *repository.PointsRepository) *UserService {
	return &UserService{
		userRepo:   userRepo,
		agentRepo:  agentRepo,
		pointsRepo: pointsRepo,
	}
}

// RegisterInput 注册输入
type RegisterInput struct {
	Username string `json:"username" binding:"required,min=1,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterOutput 注册输出
type RegisterOutput struct {
	UserID  int64  `json:"user_id"`
	AgentID int64  `json:"agent_id"`
	Token   string `json:"token"`
}

// LoginInput 登录输入
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// CreateAgentInput 创建 Agent 输入
type CreateAgentInput struct {
	Name      string  `json:"name" binding:"required,min=1,max=50"`
	AvatarURL *string `json:"avatar_url"`
	Bio       *string `json:"bio"`
}

// UpdateAgentInput 更新 Agent 输入
type UpdateAgentInput struct {
	AvatarURL *string `json:"avatar_url"`
	Bio       *string `json:"bio"`
}

// Register 用户注册：仅创建 User，不自动创建 Agent
func (s *UserService) Register(ctx context.Context, in RegisterInput, jwtSecret []byte, jwtExpireHours int) (*RegisterOutput, error) {
	// 检查邮箱是否已存在
	exist, err := s.userRepo.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		return nil, ErrUserExists
	}
	// 检查用户名是否已存在
	exist, err = s.userRepo.GetByUsername(ctx, in.Username)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		return nil, ErrUserExists
	}

	hash, err := hashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	u := &model.User{
		Username:     in.Username,
		Email:        in.Email,
		PasswordHash: hash,
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	// agentID=0 表示用户尚未创建 Agent，需调用 POST /agents 创建
	token, err := generateToken(jwtSecret, u.ID, 0, jwtExpireHours)
	if err != nil {
		return nil, err
	}

	return &RegisterOutput{UserID: u.ID, AgentID: 0, Token: token}, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, in LoginInput, jwtSecret []byte, jwtExpireHours int) (string, error) {
	u, err := s.userRepo.GetByEmail(ctx, in.Email)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", ErrInvalidCredentials
	}
	if !checkPassword(u.PasswordHash, in.Password) {
		return "", ErrInvalidCredentials
	}

	// 尝试获取 Agent，若尚未创建则 agentID=0
	var agentID int64
	a, err := s.agentRepo.GetByUserID(ctx, u.ID)
	if err != nil {
		return "", err
	}
	if a != nil {
		agentID = a.ID
	}

	return generateToken(jwtSecret, u.ID, agentID, jwtExpireHours)
}

// CreateAgent 创建 Agent（用户尚未拥有 Agent 时调用）
func (s *UserService) CreateAgent(ctx context.Context, userID int64, in CreateAgentInput, jwtSecret []byte, jwtExpireHours int) (*model.Agent, string, error) {
	// 检查用户是否已有 Agent
	exist, err := s.agentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if exist != nil {
		return nil, "", ErrAgentExists
	}
	// 检查 Agent 名称是否已被占用
	agentExist, err := s.agentRepo.GetByName(ctx, in.Name)
	if err != nil {
		return nil, "", err
	}
	if agentExist != nil {
		return nil, "", ErrAgentNameTaken
	}

	a := &model.Agent{
		UserID:  userID,
		Name:    in.Name,
		AvatarURL: in.AvatarURL,
		Bio:     in.Bio,
	}
	if err := s.agentRepo.Create(ctx, a); err != nil {
		return nil, "", err
	}

	token, err := generateToken(jwtSecret, userID, a.ID, jwtExpireHours)
	if err != nil {
		return nil, "", err
	}
	return a, token, nil
}

// GetAgentByName 获取 Agent 公开信息（含人类所有者）
func (s *UserService) GetAgentByName(ctx context.Context, name string) (*model.Agent, error) {
	return s.agentRepo.GetByNameWithUser(ctx, name)
}

// UpdateAgent 更新当前用户的 Agent
func (s *UserService) UpdateAgent(ctx context.Context, userID int64, in UpdateAgentInput) (*model.Agent, error) {
	a, err := s.agentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, nil
	}
	if in.AvatarURL != nil {
		a.AvatarURL = in.AvatarURL
	}
	if in.Bio != nil {
		a.Bio = in.Bio
	}
	if err := s.agentRepo.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *UserService) Health(ctx context.Context) error {
	return s.userRepo.Ping(ctx)
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func generateToken(secret []byte, userID, agentID int64, expireHours int) (string, error) {
	return jwt.Generate(secret, userID, agentID, expireHours)
}
