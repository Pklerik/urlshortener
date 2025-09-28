// Package links provide all business logic for links shortening app.
package links

import (
	"context"

	//nolint
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/samborkent/uuidv7"
)

// BaseLinkService - structure for service repository realization.
type BaseLinkService struct {
	linksRepo repository.LinksStorager
}

// NewLinksService - provide instance of service.
func NewLinksService(repo repository.LinksStorager) *BaseLinkService {
	return &BaseLinkService{linksRepo: repo}
}

// RegisterLinks - register the Link with provided longURL.
func (ls *BaseLinkService) RegisterLinks(ctx context.Context, longURLs []string, userID model.UserID) ([]model.LinkData, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("RegisterLink context error: %w", ctx.Err())
	}

	logger.Sugar.Infof("Long urls to shorten: %v", longURLs)
	lds := make([]model.LinkData, 0, len(longURLs))

	for _, longURL := range longURLs {
		shortURL, err := ls.cutURL(ctx, longURL)
		if err != nil {
			return lds, fmt.Errorf("(ls *LinkService) RegisterLink: %w", err)
		}

		lds = append(lds, model.LinkData{
			UUID:     model.LinkUUIDv7(uuidv7.New().String()),
			ShortURL: shortURL,
			LongURL:  longURL,
			UserID:   userID,
		})
	}

	lds, err := ls.linksRepo.SetLinks(ctx, lds)
	if err != nil && !errors.Is(err, repository.ErrExistingLink) {
		return lds, fmt.Errorf("(ls *LinkService) RegisterLink: %w", err)
	}

	if errors.Is(err, repository.ErrExistingLink) {
		return lds, repository.ErrExistingLink
	}

	return lds, nil
}

// cutURL - provide shortURl based on Long.
func (ls *BaseLinkService) cutURL(_ context.Context, longURL string) (string, error) {
	//nolint
	h := sha256.New()

	_, err := io.WriteString(h, longURL)
	if err != nil {
		return "", fmt.Errorf("(ls *BaseLinkService) cutURL: %w", err)
	}

	shortURL := fmt.Sprintf("%x", h.Sum(nil))[:8]

	return shortURL, nil
}

// GetShort - provide model.LinkData and error
// If shortURL is absent returns err.
func (ls *BaseLinkService) GetShort(ctx context.Context, shortURL string) (model.LinkData, error) {
	ld, err := ls.linksRepo.FindShort(ctx, shortURL)
	if err != nil {
		return ld, fmt.Errorf("(ls *LinkService) GetShort: %w", err)
	}

	return ld, nil
}

// PingDB - provide error if DB is not accessed.
func (ls *BaseLinkService) PingDB(ctx context.Context) error {
	if err := ls.linksRepo.PingDB(ctx); err != nil {
		return fmt.Errorf("PingDB error: %w", err)
	}

	return nil
}

// ProvideUserLinks provide user links by userID.
func (ls *BaseLinkService) ProvideUserLinks(ctx context.Context, userID model.UserID) ([]model.LinkData, error) {
	lds, err := ls.linksRepo.SelectUserLinks(ctx, userID)
	if err != nil {
		return lds, fmt.Errorf("(ls *LinkService) ProvideUserLinks: %w", err)
	}

	if len(lds) == 0 {
		return lds, repository.ErrNotFoundLink
	}

	return lds, nil
}

func (ls *BaseLinkService) MarkAsDeleted(ctx context.Context, userID model.UserID, shortLinks model.ShortUrls) (int, error) {
	return 0, nil
}

// // DeleteUserLinks deletes user links by shortUrls.
// // Returns num of deleted links and error.
// func (ls *BaseLinkService) DeleteUserLinks(ctx context.Context, shortLinks *model.ShortUrls) (int, error) {

// 	// signal chanel for goroutines closure.
// 	doneCh := make(chan struct{})
// 	// close if service done working
// 	defer close(doneCh)

// 	return nil
// }

// generator функция из предыдущего примера, делает то же, что и делала
func deletionLinksGenerator(doneCh chan struct{}, shortLinks *model.ShortUrls) chan string {
	inputCh := make(chan string)

	go func() {
		defer close(inputCh)

		for _, data := range *shortLinks {
			select {
			case <-doneCh:
				return
			case inputCh <- data:
			}
		}
	}()

	return inputCh
}

// func funInDeletionLinks(doneCh chan struct{}, shortLinks *model.ShortUrls) chan error {
// 	resErr := make(chan error)

// 	// понадобится для ожидания всех горутин
// 	var wg sync.WaitGroup

// 	for
// 	go func() {
// 		// ждём завершения всех горутин
// 		wg.Wait()
// 		// когда все горутины завершились, закрываем результирующий канал
// 		close(resErr)
// 	}()
// 	return resErr
// }
