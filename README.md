# gitlab-analyzer
gitlab project commits analyze tool

## config guide
example:
```
host: example.gitlab.com
token: $access_token
projects:
  - project1
  - project2
author: Gerrard-YNWA
from: 2021-01-30
to: 2022-01-30
```

* **host** is the gitlab instance allows to integrate with http api.
* **token** access_token with read_repository permission.
* **projects** project list wait for analyze, if not specified this tool will analyze all the projects of the access_token.
* **author** filter and only analyze specified author's commits, if not specified this tool will analyze all the authors.
**from** and **to**  time range, default scan all the history

sample:
```
go run main.go --config config.yaml

Using config file: config.yaml
Repo: project1, Commits:3
Detail:
[
	{
		"name": "Gerrard-YNWA",
		"email": "gyc.ssdut@gmail.com",
		"stats": {
			"additions": 30,
			"deletions": 3,
			"total": 27
		},
		"count": 3
	}
]
Repo: project2, Commits:3
Detail:
[
	{
		"name": "Gerrard-YNWA",
		"email": "gyc.ssdut@gmail.com",
		"stats": {
			"additions": 307,
			"deletions": 26,
			"total": 281
		},
		"count": 3
	}
]
Gitlab: 6 Commits on 2 Repos.
```
