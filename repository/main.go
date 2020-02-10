package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/anchorfree/github-terraform-exporters/pkg/repository"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	perPage = 50
)

func main() {

	tPath := flag.String("template", "templates/repo.tpl", "Template to render repos from")
	org := flag.String("org", "AnchorFree", "GitHub Organisation")
	repoType := flag.String("type", "public", "Limit by repo type (public, private)")
	fast := flag.Bool("fast", false, "Don't run per repo additional query, some parameters are not passed otherwise")
	out := flag.String("out", "", "Path to output files to in the format: repository.<repo.name>.tf. If not specified output will be printed to standard out.")

	flag.Parse()

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repoConfig := repository.ListConfig{
		ListOptions:  github.ListOptions{PerPage: perPage},
		Type:         *repoType,
		Organization: *org,
		Fast:         *fast,
	}

	repos := make(chan *github.Repository, perPage)
	go func() {
		err := repository.List(client, repos, repoConfig)
		if err != nil {
			panic(err)
		}
		close(repos)
	}()

	t := template.Must(template.ParseFiles(*tPath))
	for repo := range repos {
		var filename = fmt.Sprintf("%s/repository.%s.tf", *out, *repo.Name)

		if len(*out) > 0 {

			if _, err := os.Stat(*out); os.IsNotExist(err) {
				err := os.Mkdir(*out, 0755)
				if err != nil {
					panic(err)
				}
			}

			f, err := os.Create(filename)
			if err != nil {
				panic(err)
			}

			err = t.Execute(f, repo)
			if err != nil {
				panic(err)
			}

			err = f.Close()
			if err != nil {
				panic(err)
			}

		} else {

			err := t.Execute(os.Stdout, repo)
			if err != nil {
				panic(err)
			}

		}

	}
}
