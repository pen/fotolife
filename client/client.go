package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/jesse0michael/errgroup"
	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/net/publicsuffix"
)

const (
	BaseURI = "https://f.hatena.ne.jp/"
)

type Client struct {
	httpClient *http.Client
}

func New() *Client {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	return &Client{
		httpClient: &http.Client{
			Jar: jar,
		},
	}
}

func (c *Client) Login(ctx context.Context, id, password string) error {
	v := url.Values{}
	v.Set("name", id)
	v.Set("password", password)

	resp, err := ctxhttp.PostForm(ctx, c.httpClient, "https://www.hatena.ne.jp/login", v)
	if err != nil {
		return fmt.Errorf("on PostForm(): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("on PostForm(): status != 200")
	}

	cookies := resp.Header.Values("Set-Cookie")
	if len(cookies) == 0 {
		return fmt.Errorf("on PostForm(): no cookie")
	}

	return nil
}

func (c *Client) GetFolders(ctx context.Context, id string) ([]string, error) {
	resp, err := ctxhttp.Get(ctx, c.httpClient, BaseURI+id+"/")
	if err != nil {
		return nil, fmt.Errorf("on Get(): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("on Get(): status = %d", resp.StatusCode)
	}

	html, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("on NewDocumentFromReader(): %w", err)
	}

	var folders []string

	html.Find("ul#sidebar-folder-list li").Each(func(_ int, li *goquery.Selection) {
		href, exists := li.Find("a").Attr("href")
		if !exists || len(href) < 2 {
			return
		}

		i := strings.LastIndex(href[:len(href)-1], "/")
		if i <= 0 {
			return
		}

		folders = append(folders, href[i+1:len(href)-1])
	})

	return folders, nil
}

func (c *Client) GetViews(ctx context.Context, id string, folder string, para int) ([]string, error) {
	uri := BaseURI + id + "/"
	if folder != "" {
		uri += folder + "/"
	}

	views, lastPage, err := c.getViews(ctx, uri, 1)
	if err != nil {
		return nil, err
	}

	if para <= 0 {
		para = 1
	}

	if para > 5 {
		para = 5
	}

	eg, egCtx := errgroup.WithContext(ctx, para)
	mutex := sync.Mutex{}

	for page := 2; page <= lastPage; page++ {
		page := page

		eg.Go(func() error {
			moreViews, _, err := c.getViews(egCtx, uri, page)
			if err != nil {
				return err
			}

			mutex.Lock()
			views = append(views, moreViews...)
			mutex.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("on goroutine: %w", err)
	}

	return views, nil
}

func (c *Client) getViews(ctx context.Context, uri string, page int) ([]string, int, error) {
	if page > 1 {
		uri += fmt.Sprintf("?page=%d", page)
	}

	resp, err := ctxhttp.Get(ctx, c.httpClient, uri)
	if err != nil {
		return nil, 0, fmt.Errorf("on Get(): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("on Get(): status = %d", resp.StatusCode)
	}

	html, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("on NewDocumentFromReader(): %w", err)
	}

	var lastPage int

	if page <= 1 {
		a := html.Find("div.pager a:nth-last-child(2)").First()
		lastPage, err = strconv.Atoi(a.Text())

		if err != nil {
			lastPage = 1
		}
	}

	var views []string

	html.Find("ul.fotolist a").Each(func(_ int, a *goquery.Selection) {
		path, exists := a.Attr("href")
		if !exists {
			return
		}

		i := strings.LastIndex(path, "/")
		if i < 0 {
			return
		}

		views = append(views, path[i+1:])
	})

	return views, lastPage, nil
}

func (c *Client) GetPhotoURI(ctx context.Context, id string, view string) (string, error) {
	resp, err := ctxhttp.Get(ctx, c.httpClient, BaseURI+id+"/"+view)
	if err != nil {
		return "", fmt.Errorf("on Get(): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("on Get(): status = %d", resp.StatusCode)
	}

	html, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("on NewDocumentFromReader(): %w", err)
	}

	img := html.Find("div#foto-body img")

	src, exists := img.Attr("src")
	if !exists {
		return "", fmt.Errorf("on parsing page")
	}

	return src, nil
}
