// Package links provide all business logic for links shortening app.
package links

import (
	"context"
	"runtime"
	"slices"
	"sync"

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
	repo      repository.LinksRepository
	secretKey string
}

// NewLinksService - provide instance of service.
func NewLinksService(repo repository.LinksRepository, secretKey string) *BaseLinkService {
	return &BaseLinkService{repo: repo, secretKey: secretKey}
}

// RegisterLinks - register the Link with provided longURL.
func (ls *BaseLinkService) RegisterLinks(ctx context.Context, longURLs []string, userID model.UserID) ([]model.LinkData, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("RegisterLink context error: %w", ctx.Err())
	}

	user, err := ls.repo.CreateUser(ctx, userID)
	if err != nil {
		return []model.LinkData{}, fmt.Errorf("(ls *LinkService) RegisterLink: %w", err)
	}

	logger.Sugar.Infof("Long urls to shorten: %v", longURLs)
	lds := make([]model.LinkData, 0, len(longURLs))

	for _, longURL := range longURLs {
		shortURL, err := ls.cutURL(ctx, longURL)
		if err != nil {
			return lds, fmt.Errorf("(ls *LinkService) RegisterLink: %w", err)
		}

		lds = append(lds, model.LinkData{
			UUID:     model.UUIDv7(uuidv7.New().String()),
			ShortURL: shortURL,
			LongURL:  longURL,
			UserID:   user.ID,
		})
	}

	lds, err = ls.repo.SetLinks(ctx, lds)
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
	ld, err := ls.repo.FindShort(ctx, shortURL)
	if err != nil {
		return ld, fmt.Errorf("(ls *LinkService) GetShort: %w", err)
	}

	return ld, nil
}

// PingDB - provide error if DB is not accessed.
func (ls *BaseLinkService) PingDB(ctx context.Context) error {
	if err := ls.repo.PingDB(ctx); err != nil {
		return fmt.Errorf("PingDB error: %w", err)
	}

	return nil
}

// ProvideUserLinks provide user links by userID.
func (ls *BaseLinkService) ProvideUserLinks(ctx context.Context, userID model.UserID) ([]model.LinkData, error) {
	lds, err := ls.repo.SelectUserLinks(ctx, userID)
	if err != nil {
		return lds, fmt.Errorf("(ls *LinkService) ProvideUserLinks: %w", err)
	}

	if len(lds) == 0 {
		return lds, repository.ErrNotFoundLink
	}

	return lds, nil
}

// MarkAsDeleted - mark links as is_deleted.
func (ls *BaseLinkService) MarkAsDeleted(ctx context.Context, userID model.UserID, shortLinks model.ShortUrls) error {
	userLinks, err := ls.repo.SelectUserLinks(ctx, userID)
	if err != nil {
		return fmt.Errorf("MarkAsDeleted: %w", err)
	}

	// chan with input data
	inputCh := deletionLinksGenerator(ctx, userLinks)

	// slice of channels for parallel work size of gomaxpoc.
	channels := fanOutDeletionLinks(ctx, inputCh, shortLinks)

	// collect all channels to 1.
	addResultCh := funInDeletionLinks(ctx, channels...)

	err = ls.repo.BatchMarkAsDeleted(ctx, addResultCh)
	if err != nil {
		return fmt.Errorf("MarkAsDeleted: %w", err)
	}

	return nil
}

func linkForDeletion(ctx context.Context, userLinkCh chan model.LinkData, inputLinks model.ShortUrls) chan model.LinkData {
	linksForDeletionCh := make(chan model.LinkData)

	go func() {
		// откладываем сообщение о том, что горутина завершилась
		defer close(linksForDeletionCh)

		for ul := range userLinkCh {
			if !slices.Contains(inputLinks, ul.ShortURL) {
				continue
			}

			select {
			case <-ctx.Done():
				return
			case linksForDeletionCh <- ul:
			}
		}
	}()

	return linksForDeletionCh
}

// generator функция из предыдущего примера, делает то же, что и делала.
func deletionLinksGenerator(ctx context.Context, links []model.LinkData) chan model.LinkData {
	inputCh := make(chan model.LinkData)

	go func() {
		defer close(inputCh)

		for _, data := range links {
			select {
			case <-ctx.Done():
				return
			case inputCh <- data:
			}
		}
	}()

	return inputCh
}

// fanOut принимает канал данных, порождает 10 горутин.
func fanOutDeletionLinks(ctx context.Context, inputCh chan model.LinkData, inputLinks model.ShortUrls) []chan model.LinkData {
	// количество горутин add
	numWorkers := runtime.GOMAXPROCS(0)
	// каналы, в которые отправляются результаты
	channels := make([]chan model.LinkData, numWorkers)

	for i := 0; i < numWorkers; i++ {
		// получаем канал из горутины add
		addResultCh := linkForDeletion(ctx, inputCh, inputLinks)
		// отправляем его в слайс каналов
		channels[i] = addResultCh
	}
	// возвращаем слайс каналов
	return channels
}

// fanIn объединяет несколько каналов resultChs в один.
func funInDeletionLinks(ctx context.Context, resultChs ...chan model.LinkData) chan model.LinkData {
	// конечный выходной канал в который отправляем данные из всех каналов из слайса, назовём его результирующим
	finalCh := make(chan model.LinkData)

	// понадобится для ожидания всех горутин
	var wg sync.WaitGroup

	// перебираем все входящие каналы
	for _, ch := range resultChs {
		// в горутину передавать переменную цикла нельзя, поэтому делаем так
		chClosure := ch

		// инкрементируем счётчик горутин, которые нужно подождать
		wg.Add(1)

		go func() {
			// откладываем сообщение о том, что горутина завершилась
			defer wg.Done()

			// получаем данные из канала
			for data := range chClosure {
				select {
				// выходим из горутины, если канал закрылся
				case <-ctx.Done():
					return
				// если не закрылся, отправляем данные в конечный выходной канал
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		// ждём завершения всех горутин
		wg.Wait()
		// когда все горутины завершились, закрываем результирующий канал
		close(finalCh)
	}()

	// возвращаем результирующий канал
	return finalCh
}

// GetSecret provide secret key from service.
func (ls *BaseLinkService) GetSecret(name string) (string, bool) {
	switch name {
	case "SECRET_KEY":
		return ls.secretKey, true
	default:
		return "", false
	}
}

// GetStats - provide aggregated statistics about shortened links and users.
func (ls *BaseLinkService) GetStats(ctx context.Context) (model.Stats, error) {
	stats, err := ls.repo.GetStats(ctx)
	if err != nil {
		return model.Stats{}, fmt.Errorf("GetStats: %w", err)
	}

	return stats, nil
}
