package app

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/config/mocks"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/golang/mock/gomock"
)

func init() {
	// Disable logger output for tests
	logger.Initialize("INFO")
}

func TestStartApp_SuccessfulStartup(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockParser := mocks.NewMockStartupFlagsParser(ctrl)
	defer ctrl.Finish()
	mockParser.EXPECT().GetServerAddress().Return(config.Address{Port: 8080}).AnyTimes()
	mockParser.EXPECT().GetDatabaseConf().Return(nil, errors.New("")).AnyTimes()
	mockParser.EXPECT().GetLocalStorage().Return("").AnyTimes()
	mockParser.EXPECT().GetTimeout().Return(5 * time.Second).AnyTimes()
	mockParser.EXPECT().GetSecretKey().Return("secret").AnyTimes()
	mockParser.EXPECT().GetTLS().Return(false).AnyTimes()
	mockParser.EXPECT().GetTrustedCIDR().Return("127.0.0.1/8").AnyTimes()

	go func() {
		time.Sleep(100 * time.Millisecond)
		syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}()

	StartApp(mockParser)
}

func TestStartApp_InvalidPort(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockParser := mocks.NewMockStartupFlagsParser(ctrl)
	defer ctrl.Finish()
	mockParser.EXPECT().GetServerAddress().Return(config.Address{Port: -1}).AnyTimes()
	mockParser.EXPECT().GetDatabaseConf().Return(nil, errors.New("")).AnyTimes()
	mockParser.EXPECT().GetLocalStorage().Return("").AnyTimes()
	mockParser.EXPECT().GetTimeout().Return(5 * time.Second).AnyTimes()
	mockParser.EXPECT().GetSecretKey().Return("secret").AnyTimes()
	mockParser.EXPECT().GetTLS().Return(false).AnyTimes()
	mockParser.EXPECT().GetTrustedCIDR().Return("127.0.0.1/8").AnyTimes()

	StartApp(mockParser)
}
