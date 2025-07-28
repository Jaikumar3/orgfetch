# orgfetch

A powerful Go CLI tool to fetch and report organization/group members and repositories from **GitHub** or **GitLab**. Supports advanced filtering, flexible output, and repo downloading.

## Features
- Fetch organization (GitHub) or group (GitLab) members and repositories
- Filter by member, repo type, and fork status
- Print colored output and repo URLs
- Download repositories (with size limit)
- Output results to file
- Read multiple orgs/groups from a file (pass filename to --orgname)
- Print only repo URLs (`--urls-only`) or only usernames (`--usernames-only`)
- Flexible repo type selection: org, member, both
- All flags have short forms for usability
- Select provider: `--provider github|gitlab`

## Installation

1. Clone the repository and build:
   ```sh
   git clone <repo-url>
   cd <repo-folder>
   go build -o orgfetch main.go github.go gitlab.go fetcher.go
   ```
2. (Optional) Move the binary to your PATH:
   ```sh
   sudo mv orgfetch /usr/local/bin/
   ```

## Usage

### Basic Examples

Fetch all GitHub org repos (default, forks excluded):
```
./orgfetch --provider github --token <TOKEN> --orgname <ORG>
```

Fetch all GitLab group repos:
```
./orgfetch --provider gitlab --token <TOKEN> --orgname <GROUP>
```

Fetch all org/group repos, including forks:
```
./orgfetch --provider github --token <TOKEN> --orgname <ORG> --include-forks
./orgfetch --provider gitlab --token <TOKEN> --orgname <GROUP> --include-forks
```

Fetch member-owned repos for all members/users:
```
./orgfetch --provider github --token <TOKEN> --orgname <ORG> --repo-type member
./orgfetch --provider gitlab --token <TOKEN> --orgname <GROUP> --repo-type member
```

Print only repo URLs:
```
./orgfetch --provider github --token <TOKEN> --orgname <ORG> --urls-only
./orgfetch --provider gitlab --token <TOKEN> --orgname <GROUP> --urls-only
```

Print only usernames of org/group members:
```
./orgfetch --provider github --token <TOKEN> --orgname <ORG> --usernames-only
./orgfetch --provider gitlab --token <TOKEN> --orgname <GROUP> --usernames-only
```

Download all org/group repos (max size 250MB):
```
./orgfetch --provider github --token <TOKEN> --orgname <ORG> --download
./orgfetch --provider gitlab --token <TOKEN> --orgname <GROUP> --download
```

Save results to a file:
```
./orgfetch --provider github --token <TOKEN> --orgname <ORG> --output results.txt
./orgfetch --provider gitlab --token <TOKEN> --orgname <GROUP> --output results.txt
```

Fetch for multiple orgs/groups listed in a file:
```
./orgfetch --provider github --token <TOKEN> --orgname orgs.txt --output results.txt
./orgfetch --provider gitlab --token <TOKEN> --orgname groups.txt --output results.txt
```

## Flags

- `--provider`, `-p`: Provider to use: github or gitlab (default: github)
- `--token`, `-t`: Personal access token (required)
- `--orgname`, `-o`: Organization (GitHub) or group (GitLab) name (required)
- `--output`, `-O`: Write results to output file
- `--include-forks`, `-f`: Include forked repositories in the output
- `--repo-type`, `-r`: Type of repositories to fetch: org, member, both (default: org)
- `--member`, `-m`: Username to fetch repos for particular user/member (only used with --repo-type member)
- `--download`, `-d`: Download all listed repositories using git clone
- `--urls-only`, `-u`: Print only repo URLs (no names)
- `--usernames-only`, `-U`: Print only usernames of organization/group members (one per line)
- `--max-size`, `-s`: Maximum repo size (MB) to clone (default: 250)

## Notes
- For GitHub, a personal access token is always required.
- For GitLab, a token is required (public-only support can be added if needed).
- For multiple orgs/groups, provide a file with one name per line to `--orgname`.
- All output modes (console, file, pipe) work for multiple orgs/groups and all flag combinations.

## Author
Jai aka hacdoc
## License
MIT

## License
MIT
