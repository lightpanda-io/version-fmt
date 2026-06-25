// Copyright 2026 Lightpanda (Selecy SAS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	exitOK   = 0
	exitFail = 1

	repository = "lightpanda-io/browser"
)

// main starts interruptable context and runs the program.
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	err := run(ctx, os.Args, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(exitFail)
	}

	os.Exit(exitOK)
}

func run(_ context.Context, args []string, stdout, stderr io.Writer) error {
	// declare runtime flag parameters.
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	var (
		release = flags.String("release", "nightly", "release tag")
		version = flags.String("version", "", "version string from `lightpanda version`")
	)

	// usage func declaration.
	exec := args[0]
	flags.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s <version.json>\n", exec)            //nolint:errcheck
		fmt.Fprintf(stderr, "\nRead, format and save versions results.\n") //nolint:errcheck
		fmt.Fprintf(stderr, "\nCommand line options:\n")                   //nolint:errcheck
		flags.PrintDefaults()
		fmt.Fprintf(stderr, "\nTo retrieve release info from GH, the program uses env var:\n") //nolint:errcheck
	}
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return errors.New("bad arguments")
	}

	file := args[0]

	// entry is one platform object in the manifest. size is a string, like Zig.
	type entry struct {
		DownloadURL string `json:"download_url"`
		Shasum      string `json:"shasum"`
		Size        string `json:"size"`
	}

	rel, err := fetchRelease(repository, *release)
	if err != nil {
		return fmt.Errorf("fetch release: %w", err)
	}

	out := map[string]any{"version": *version}
	var date time.Time
	for _, v := range rel.Assets {

		// Filter out non-expected assets
		ok := false
		for _, p := range []string{
			"lightpanda-aarch64-linux", "lightpanda-x86_64-linux",
			"lightpanda-aarch64-macos", "lightpanda-x86_64-macos",
		} {
			if v.Name == p {
				ok = true
			}
		}

		if !ok {
			continue
		}

		date = v.CreatedAt
		sha := strings.TrimPrefix(v.Digest, "sha256:")
		p := strings.TrimPrefix(v.Name, "lightpanda-")

		if sha == "" {
			return fmt.Errorf("missing release asset or digest for lightpanda-%s", p)
		}

		out[p] = entry{
			DownloadURL: v.DownloadURL,
			Shasum:      sha,
			Size:        fmt.Sprint(v.Size),
		}
	}
	out["date"] = date.Format("2006-01-02")

	index, err := readIndex(file)
	if err != nil {
		return fmt.Errorf("read index: %w", err)
	}

	raw, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("marshal entry: %w", err)
	}
	index[*release] = raw

	buf, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	if _, err := io.Copy(stdout, bytes.NewReader(buf)); err != nil {
		return fmt.Errorf("output: %w", err)
	}

	return nil
}

type Release struct {
	Tag    string `json:"tag"`
	Assets []struct {
		Name        string    `json:"name"`
		CreatedAt   time.Time `json:"created_at"`
		DownloadURL string    `json:"browser_download_url"`
		Size        int64     `json:"size"`
		Digest      string    `json:"digest"`
	} `json:"assets"`
}

func fetchRelease(repo, release string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, release)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch release: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch release %s: HTTP %d", url, resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)

	var rel Release
	if err := dec.Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}

	return &rel, nil
}

func readIndex(path string) (map[string]json.RawMessage, error) {
	index := map[string]json.RawMessage{}
	if path == "" {
		return index, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return index, nil
		}
		return nil, fmt.Errorf("read %s: %v", path, err)
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return index, nil
	}
	if err := json.Unmarshal(b, &index); err != nil {
		return nil, fmt.Errorf("decode %s: %v", path, err)
	}
	return index, nil
}
