# GoGitlabber

This project is inspired from the python application called gitlabber (https://github.com/ezbz/gitlabber).
It is mainly to learn Golang. But also to make something that specifically solves my problem. 😆

It is definitely not as feature-rich as the original project... 😬

The program can clone and pull all repositories you have access to on a selfhosted or SaaS provided Gitlab or Gitea
server.
It only supports the HTTP access method.

It will pull the repositories in a tree like structure same as on Gitlab or Gitea.

```
root [http://gitlab.example.com]
├── group1 [/group1]
│   └── subgroup1 [/group1/subgroup1]
│       └── project1 [/group1/subgroup1/project1]
└── group2 [/group2]
    ├── subgroup1 [/group2/subgroup1]
    │   └── project2 [/group2/subgroup1/project2]
    ├── subgroup2 [/group2/subgroup2]
    └── subgroup3 [/group2/subgroup3]
```

## Config file

GitLab:

```yaml
# ~/.config/gogitlabber/gitlab.example.com.yaml
debug: false
concurrency: 15
git_host: "gitlab.example.net"
git_token: "glpat-"
git_backend: "gitlab"
include_archived: "excluded"
destination: "$HOME/Documents"
```


## Usage

```bash
gogitlabber -config=~/.config/gogitlabber/gitlab.example.com.yaml
```


## Access Token Permissions

### Gitea

Make sure the Gitea Access Token has at least the following permissions:
- user - read
- repository - read

### Gitlab

Make sure the Gitlab Access Token has the `api` scope.
