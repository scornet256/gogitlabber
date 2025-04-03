# GoGitlabber

This project is inspired from the python application called gitlabber (https://github.com/ezbz/gitlabber).
It is mainly to learn Golang. But also to make something that specifically
solves my problem. 😆

It is definitely not as feature-rich as the original project... 😬

The program can clone and pull all repositories you have access to on a selfhosted or SaaS provided Gitlab server.
It only supports the HTTP access method.

It will pull the repositories in a tree like structure same as on Gitlab.
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

# Usage

```
Usage of gogitlabber:
  -archived string
        To include archived repositories (any|excluded|exclusive)
          example: -archived=any
        env = GOGITLABBER_ARCHIVED
         (default "excluded")

  -concurrency int
        Specify repository concurrency
          example: -concurrency=15
        env = GOGITLABBER_CONCURRENCY
         (default 15)

  -debug
        Toggle debug mode
         example: -debug=true
        env = GOGITLABBER_DEBUG
         (default false)

  -destination string
        Specify where to check the repositories out
          example: -destination=$HOME/repos
        env = GOGITLABBER_DESTINATION
         (default "$HOME/Documents")

  -gitlab-api-token string
        Specify GitLab API token
          example: -gitlab-api=glpat-xxxx
        env = GITLAB_API_TOKEN
         (default "")

  -gitlab-url string
        Specify GitLab host
          example: -gitlab-url=gitlab.example.com
        env = GITLAB_URL
         (default "gitlab.com")
```
