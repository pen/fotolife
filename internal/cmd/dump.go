package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/remeh/sizedwaitgroup"
	"golang.org/x/net/context/ctxhttp"

	"github.com/pen/fotolife/client"
)

//nolint:maligned
type Dump struct {
	TargetID string   `kong:"arg,help='Hatena-ID to dump'"`
	Folders  []string `kong:"arg,optional,help='Specify folders'"`

	Top    bool   `kong:"short='t',help='Include top fotos'"`
	Dir    string `kong:"short='d',help='Directory to save'"`
	Update bool   `kong:"short='u',help='Skip if already downloaded'"`
	DryRun bool   `kong:"short='n',help='Don\\'t download (use with -v)'"`
	Para   int    `kong:"hidden"`
}

func (d *Dump) Run(o *Options) error {
	ctx := context.Background()
	c := client.New()

	if o.Password != "" {
		id := d.TargetID
		if o.LoginID != "" {
			id = o.LoginID
		}

		o.Debugf("login as %s", id)

		if err := c.Login(ctx, id, o.Password); err != nil {
			return fmt.Errorf("on Login(): %w", err)
		}
	}

	return d.processFolders(ctx, c, o)
}

func (d *Dump) processFolders(ctx context.Context, c *client.Client, o *Options) error {
	folders, err := c.GetFolders(ctx, d.TargetID)
	if err != nil {
		return fmt.Errorf("on GetFolders(): %w", err)
	}

	o.Debugf("remote folders(%d): %v", len(folders), folders)

	if len(d.Folders) > 0 {
	L:
		for _, optFolder := range d.Folders {
			for _, folder := range folders {
				if folder == optFolder {
					continue L
				}
			}

			return fmt.Errorf("no such folder: %s", optFolder)
		}

		folders = d.Folders
	}

	o.Debugf("dump folders(%d): %v", len(folders), folders)

	if d.Top {
		folders = append(folders, "")
	}

	for _, folder := range folders {
		views, err := c.GetViews(ctx, d.TargetID, folder, d.Para)
		if err != nil {
			o.Debugf("GetViews(%s): %v", folder, err)
			continue
		}

		o.Debugf("views of [%s]: %d", folder, len(views))

		if len(views) == 0 {
			continue
		}

		dir, err := d.makeDir(folder, o)
		if err != nil {
			o.Debugf("makeDir(%s): %v", folder, err)
			continue
		}

		d.processViews(ctx, views, dir, c, o)
	}

	return nil
}

func (d *Dump) processViews(ctx context.Context, views []string, dir string, c *client.Client, o *Options) {
	para := d.Para
	if para <= 0 {
		para = 1
	}

	if para > 5 {
		para = 5
	}

	o.Debugf("goroutines: %d", para)

	swg := sizedwaitgroup.New(para)

	for _, view := range views {
		swg.Add()

		go func(v string) {
			defer swg.Done()

			uri, err := c.GetPhotoURI(ctx, d.TargetID, v)
			if err != nil {
				o.Debugf("failed to get uri of %s", v)
				return
			}

			err = d.download(ctx, http.DefaultClient, uri, dir, o)
			if err != nil {
				o.Debugf("failed to download: %s %s", dir, uri)
				return
			}
		}(view)
	}

	swg.Wait()
}

func (d *Dump) makeDir(folder string, o *Options) (string, error) {
	dir := d.Dir
	if dir == "" {
		dir = d.TargetID
	}

	if dir != "" && !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	dir += folder
	if dir == "" || dir == "/" {
		return dir, nil
	}

	o.Verbosef(`mkdir -p "%s"`, dir)

	if d.DryRun {
		return dir, nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("on MkdirAll(): %w", err)
	}

	return dir, nil
}

func (d *Dump) download(ctx context.Context, httpClient *http.Client, uri, dir string, o *Options) error {
	index := strings.LastIndex(uri, "/")
	if index < 0 {
		return fmt.Errorf("on making filename from uri: %s", uri)
	}

	if dir != "" && !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	path := dir + uri[index+1:]

	if d.Update {
		if d.canSkip(ctx, httpClient, uri, path) {
			o.Debugf("skip: %s", path)
			return nil
		}
	}

	o.Verbosef(`curl -s -o "%s" "%s"`, path, uri)

	if d.DryRun {
		return nil
	}

	resp, err := ctxhttp.Get(ctx, httpClient, uri)
	if err != nil {
		return fmt.Errorf("on Get(): %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("on Create(): %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("on Copy(): %w", err)
	}

	return nil
}

func (d *Dump) canSkip(ctx context.Context, httpClient *http.Client, uri, path string) bool {
	resp, err := ctxhttp.Head(ctx, httpClient, uri)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	lengths := resp.Header.Values("Content-Length")
	if len(lengths) == 0 {
		return false
	}

	length, err := strconv.Atoi(lengths[0])
	if err != nil {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if info.Size() != int64(length) {
		return false
	}

	return true
}
