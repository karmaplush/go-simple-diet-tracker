package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	grpclogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	grpcauthv1 "github.com/karmaplush/proto-contracts/gen/go/grpcauth"
	"github.com/karmaplush/simple-diet-tracker/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	api   grpcauthv1.AuthClient
	log   *slog.Logger
	appId int32
}

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user exists")
	ErrGRPCUnexpected  = errors.New("unexpected GRCP error")
	ErrInvalidArgument = errors.New("invalid argument")
)

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg *config.Config,
) (*Client, error) {
	const op = "grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(cfg.Clients.GRPCAuth.RetriesCount)),
		grpcretry.WithPerRetryTimeout(cfg.Clients.GRPCAuth.Timeout),
	}

	logOpts := []grpclogging.Option{
		grpclogging.WithLogOnEvents(grpclogging.PayloadReceived, grpclogging.PayloadSent),
	}

	cc, err := grpc.DialContext(
		ctx,
		cfg.Clients.GRPCAuth.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcretry.UnaryClientInterceptor(retryOpts...),
			grpclogging.UnaryClientInterceptor(InterseptorLogger(log), logOpts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api:   grpcauthv1.NewAuthClient(cc),
		log:   log,
		appId: cfg.AppId,
	}, nil
}

// Adapter slog logger -> interceptor logger
func InterseptorLogger(l *slog.Logger) grpclogging.Logger {
	return grpclogging.LoggerFunc(
		func(ctx context.Context, level grpclogging.Level, msg string, fields ...any) {
			l.Log(ctx, slog.Level(level), msg, fields...)
		},
	)

}

func (c *Client) Login(
	ctx context.Context,
	email string,
	password string,
) (token string, userId int64, err error) {
	const op = "grpc.Login"

	resp, err := c.api.Login(
		ctx,
		&grpcauthv1.LoginRequest{Email: email, Password: password, ServiceId: c.appId},
	)

	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return "", 0, fmt.Errorf("%s: %w", op, ErrInvalidArgument)
			case codes.Unauthenticated:
				return "", 0, fmt.Errorf("%s: %w", op, ErrUserNotFound)
			default:
				return "", 0, fmt.Errorf("%s: %w", op, ErrGRPCUnexpected)
			}
		} else {
			return "", 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	return resp.GetToken(), resp.GetUserId(), nil
}

func (c *Client) Registration(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const op = "grpc.Registration"

	resp, err := c.api.Register(
		ctx,
		&grpcauthv1.RegisterRequest{Email: email, Password: password},
	)

	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return 0, fmt.Errorf("%s: %w", op, ErrInvalidArgument)
			case codes.AlreadyExists:
				return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
			default:
				return 0, fmt.Errorf("%s: %w", op, ErrGRPCUnexpected)
			}
		} else {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	return resp.GetUserId(), err
}
