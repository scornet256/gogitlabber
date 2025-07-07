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
concurrency: 15
debug: false
destination: "$HOME/Documents"
git_backend: "gitlab"
git_host: "gitlab.example.com"
git_token: "glpat-"
git_user_mail: "john.doe@example.com"
git_user_name: "John Doe"
include_archived: "excluded"
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
