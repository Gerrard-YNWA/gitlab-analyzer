package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Gitlab struct {
	Apikey            string
	Host              string
	BasicPath         string
	SpecifiedProjects []string
}

type Author struct {
	Name  string      `json:"name"`
	Email string      `json:"email"`
	Stats CommitStats `json:"stats"`
	Count int         `json:"count"`
}

type CommitStats struct {
	Add   int `json:"additions"`
	Del   int `json:"deletions"`
	Total int `json:"total"`
}

type Commit struct {
	Id           string      `json:"id"`
	ShortId      string      `json:"short_id"`
	Title        string      `json:"title"`
	Message      string      `json:"message"`
	AuthorName   string      `json:"author_name"`
	AuthorEmail  string      `json:"author_email"`
	AuthoredDate string      `json:"authored_date"`
	Stats        CommitStats `json:"stats"`

	gitlab *Gitlab
}

type Repo struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	AuthorInfos     map[string]*Author
	Commits         []Commit
	FilteredCommits []Commit
	SpecifiedAuthor string
	From            string
	To              string

	gitlab *Gitlab
}

func normalizeAuthoredDate(date string) string {
	ss := strings.Split(date, "T")
	return ss[0]
}

func collectCommits(r *Repo, cmt Commit) {
	r.Commits = append(r.Commits, cmt)
	if !strings.HasPrefix(strings.TrimPrefix(cmt.Title, " "), "Merge branch") {
		r.FilteredCommits = append(r.FilteredCommits, cmt)
		if v, ok := r.AuthorInfos[cmt.AuthorName]; ok {
			v.Stats.Add += cmt.Stats.Add
			v.Stats.Del += cmt.Stats.Del
			v.Stats.Total = v.Stats.Add - v.Stats.Del
			v.Count += 1
		} else {
			committer := &Author{
				Name:  cmt.AuthorName,
				Email: cmt.AuthorEmail,
				Stats: cmt.Stats,
				Count: 1,
			}
			r.AuthorInfos[cmt.AuthorName] = committer
		}
	}
}

func New(host, key, path string) *Gitlab {
	return &Gitlab{
		Host:      host,
		Apikey:    key,
		BasicPath: path,
	}
}

func (g *Gitlab) WithSpecifiedProjects(projects []string) *Gitlab {
	g.SpecifiedProjects = projects
	return g
}

func (g *Gitlab) req(pat string, args ...interface{}) (*http.Response, error) {
	xargs := []interface{}{g.Host, g.BasicPath}
	xargs = append(xargs, args...)
	xargs = append(xargs, g.Apikey)
	url := fmt.Sprintf(pat, xargs...)
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(request)
}

func (g *Gitlab) FetchRepos() ([]*Repo, error) {
	page := 0
	perpage := 20
	pat := "https://%s%s/projects?page=%d&per_page=%d&private_token=%s"
	resp, err := g.req(pat, page, perpage)
	if err != nil {
		return nil, err
	}

	pages, err := strconv.Atoi(resp.Header.Get("X-Total-Pages"))
	if err != nil {
		return nil, err
	}

	var repos []*Repo
	for page = 1; page <= pages; page++ {
		resp, err := g.req(pat, page, perpage)
		if err != nil {
			return nil, err
		}
		if data, err := io.ReadAll(resp.Body); err != nil {
			return nil, err
		} else {
			var rs []Repo
			if err := json.Unmarshal(data, &rs); err != nil {
				return nil, err
			}

			for _, name := range g.SpecifiedProjects {
				for _, v := range rs {
					if v.Name == name {
						v.AuthorInfos = make(map[string]*Author)
						v.gitlab = g
						repos = append(repos, &v)
						break
					}
				}
			}
		}

		if len(repos) == len(g.SpecifiedProjects) {
			break
		}
	}

	return repos, nil
}

func (r *Repo) WithSpecifiedAuthor(author string) *Repo {
	r.SpecifiedAuthor = author
	return r
}

func (r *Repo) WithDuration(from, to string) *Repo {
	r.From = from
	r.To = to
	return r
}

func (r *Repo) FetchCommits() error {
	page := 1
	perpage := 20
	pat := "https://%s%s/projects/%d/repository/commits?page=%d&per_page=%d&private_token=%s"

	for {
		resp, err := r.gitlab.req(pat, r.Id, page, perpage)
		if err != nil {
			return err
		}

		if data, err := io.ReadAll(resp.Body); err != nil {
			return err
		} else {
			var cmts []Commit
			if err := json.Unmarshal(data, &cmts); err != nil {
				return err
			}

			for _, cmt := range cmts {
				cmtDate := normalizeAuthoredDate(cmt.AuthoredDate)
				if r.From != "" && cmtDate < r.From {
					return nil
				}

				if r.To != "" && cmtDate > r.To {
					return nil
				}

				if err := r.fetchStats(&cmt); err != nil {
					return err
				}

				if r.SpecifiedAuthor != "" && cmt.AuthorName == r.SpecifiedAuthor {
					collectCommits(r, cmt)
				} else if r.SpecifiedAuthor == "" {
					collectCommits(r, cmt)
				}
			}
		}

		next := resp.Header.Get("X-Next-Page")
		if next == "" {
			break
		}

		page, err = strconv.Atoi(next)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repo) fetchStats(commit *Commit) error {
	pat := "https://%s:%s/projects/%d/repository/commits/%s?private_token=%s"
	resp, err := r.gitlab.req(pat, r.Id, commit.Id)
	if err != nil {
		return err
	}
	if data, err := io.ReadAll(resp.Body); err != nil {
		return err
	} else {
		if err := json.Unmarshal(data, commit); err != nil {
			return err
		}
	}

	return nil
}
