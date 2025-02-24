# gogitlabber
This project is inspired from the python application called gitlabber (https://github.com/ezbz/gitlabber).
It is mainly to learn Golang. But also to make something that specifically solves my problem. :)

The program can download and update all repositories you have access to on a Gitlab server.
This works for bot gitlab.com and a selfhosted Gitlab instance. It only supports the HTTP method.

# Usage
```
Usage: gogitlabber
         --archived=(any|excluded|only)
         --destination=$HOME/Documents
         --gitlab-url=gitlab.com
         --gitlab-token=<supersecrettoken>

You can also set these environment variables:
  GOGITLABBER_ARCHIVED=(any|excluded|only)
  GOGITLABBER_DESTINATION=$HOME/Documents
  GITLAB_API_TOKEN=<supersecrettoken>
  GITLAB_URL=gitlab.com
```
