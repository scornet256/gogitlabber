# GoGitlabber

This project is inspired from the python application called gitlabber (https://github.com/ezbz/gitlabber).
It is mainly to learn Golang. But also to make something that specifically
solves my problem. 😆

It is definitely not as feature-rich as the original project... 😬

The program can clone and pull all repositories you have access to on a selfhosted or SaaS provided Gitlab or Gitea server.
It only supports the HTTP access method.

It will pull the repositories in a tree like structure same as on Gitlab or
Gitea.
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

# Example
Gitea:
```
$ ./gogitlabber -backend=gitea -destination=$HOME/Documents -git-url=gitea.example.com
-git-api-token=supersecrettoken
 100% [====================] (30/30) [4s] ...   
Summary:
 Cloned repositories: 0
 Pulled repositories: 30
 Errors: 0
```
Gitlab:
```
$ ./gogitlabber -backend=gitlab -destination=$HOME/Documents -git-url=gitlab.example.com
-git-api-token=supersecrettoken
 100% [====================] (30/30) [4s] ...   
Summary:
 Cloned repositories: 0
 Pulled repositories: 30
 Errors: 0
```

# Usage
```
Usage of gogitlabber:
  -archived string
        To include archived repositories (any|excluded|exclusive)
          example: -archived=any
        env = GOGITLABBER_ARCHIVED
         (default "excluded")

  -backend string
        Specify git backend
          example: -backend=gitlab
        env = GOGITLABBER_BACKEND

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

  -git-api-token string
        Specify API token
          example: -git-api=glpat-xxxx
        env = GIT_API_TOKEN
         (default "")

  -git-url string
        Specify Git host
          example: -git-url=gitlab.example.com
        env = GIT_URL
         (default "gitlab.com")
```

# Access Token Permissions
## Gitea
Make sure the Gitea Access Token has at least the following permissions:
- user - read
- repository - read

## Gitlab
Make sure the Gitlab Access Token has the `api` scope.
