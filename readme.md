# GoGitlabber

This project is inspired from the python application called gitlabber (https://github.com/ezbz/gitlabber).
It is mainly to learn Golang. But also to make something that specifically solves my problem. 😆

It is definitely not as feature-rich as the original project... 😬

The program can clone and pull all repositories you have access to on a selfhosted or SaaS provided Gitlab or Gitea
server.
It clones over HTTPS (token in the remote URL) or over SSH (`git_protocol: "ssh"`, no credentials in the remote URL).

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

| Key                | Required | Default           | Description                                              |
|--------------------|----------|-------------------|----------------------------------------------------------|
| `git_backend`      | yes      |                   | `gitlab` or `gitea`                                      |
| `git_token`        | yes      |                   | API / access token                                       |
| `git_host`         | no       | `gitlab.com`      | Hostname of the server                                   |
| `git_protocol`     | no       | `https`           | `https` or `ssh`, used for clone and pull                |
| `git_ssh_port`     | no       | `22`              | SSH port in the remote URL (only with `ssh`)             |
| `git_ssh_user`     | no       | `git`             | SSH user in the remote URL (only with `ssh`)             |
| `destination`      | no       | `$HOME/Documents` | Directory to clone into; `$VARS` and `~` are expanded    |
| `concurrency`      | no       | `15`              | Number of repositories processed in parallel             |
| `include_archived` | no       | `excluded`        | `any`, `excluded` or `exclusive` (only archived)         |
| `git_user_name`    | no       |                   | Sets `user.name` in each repository when set             |
| `git_user_mail`    | no       |                   | Sets `user.email` in each repository when set            |
| `debug`            | no       | `false`           | Verbose logging instead of the progress bar              |

Unknown keys are rejected, so typos are caught early.

## Usage

```bash
gogitlabber -config=~/.config/gogitlabber/gitlab.example.com.yaml
```

Without `-config`, `~/.config/gogitlabber.yaml` is used.

Repositories with local changes are skipped and listed in the summary. Only the `origin` remote is updated with the
current URL, so switching `git_protocol` moves existing clones on the next run.

## SSH

With `git_protocol: "ssh"` the remote URL is `ssh://<git_ssh_user>@<git_host>:<git_ssh_port>/<repo>.git`. Clone and
pull run through the `git` CLI, so OpenSSH applies `~/.ssh/config` (ProxyJump, IdentityFile, known hosts). SSH runs
in batch mode: keys need an agent or no passphrase, and unknown host keys fail instead of prompting. The token is
still needed to list the repositories through the API.

## Access Token Permissions

### Gitea

Make sure the Gitea Access Token has at least the following permissions:
- user - read
- repository - read

### Gitlab

Make sure the Gitlab Access Token has the `api` scope.
