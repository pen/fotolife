package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/pen/fotolife/client"

	"github.com/remeh/sizedwaitgroup"
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
	c := client.New()

	if o.Password != "" {
		id := d.TargetID
		if o.LoginID != "" {
			id = o.LoginID
		}

		o.Debugf("login as %s", id)

		err := c.Login(id, o.Password)
		if err != nil {
			return err
		}
	}

	return d.processFolders(c, o)
}

func (d *Dump) processFolders(c *client.Client, o *Options) error {
	folders, err := c.GetFolders(d.TargetID)
	if err != nil {
		return err
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
		views, err := c.GetViews(d.TargetID, folder, d.Para)
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

		d.processViews(views, dir, c, o)
	}

	return nil
}

func (d *Dump) processViews(views []string, dir string, c *client.Client, o *Options) {
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

			uri, err := c.GetPhotoURI(d.TargetID, v)
			if err != nil {
				o.Debugf("failed to get uri of %s", v)
				return
			}

			err = d.download(uri, dir, o)
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

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	return dir, nil
}

func (d *Dump) download(uri, dir string, o *Options) error {
	index := strings.LastIndex(uri, "/")
	if index < 0 {
		return fmt.Errorf("failed to make filename from uri: %s", uri)
	}

	if dir != "" && !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	path := dir + uri[index+1:]

	if d.Update {
		if d.canSkip(uri, path) {
			o.Debugf("skip: %s", path)
			return nil
		}
	}

	o.Verbosef(`curl -s -o "%s" "%s"`, path, uri)

	if d.DryRun {
		return nil
	}

	resp, err := http.Get(uri) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return err
}

func (d *Dump) canSkip(uri, path string) bool {
	resp, err := http.Head(uri) //nolint:gosec
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
