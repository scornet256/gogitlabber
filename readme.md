# gogitlabber
This project is inspired from the python application called gitlabber (https://github.com/ezbz/gitlabber).
It is mainly to learn Golang. But also to make something that specifically solves my problem. :)

The program can clone and pull all repositories you have access to on a selfhosted or SaaS provided Gitlab server.
It only supports the HTTP access method.

# Usage
```
Usage of gogitlabber:
  -archived string
        To include archived repositories (any|excluded|exclusive)
          example: -archived=any
        env = GOGITLABBER_ARCHIVED
         (default "excluded")

  -destination string
        Specify where to check the repositories out
          example: -destination=$HOME/repos
        env = GOGITLABBER_DESTINATION
         (default "$HOME/Documents")

  -gitlab-api-token string
        Specify GitLab API token
          example: -gitlab-api=glpat-xxxx
        env = GITLAB_API_TOKEN

  -gitlab-url string
        Specify GitLab host
          example: -gitlab-url=gitlab.com
        env = GITLAB_URL
         (default "gitlab.com")
```
