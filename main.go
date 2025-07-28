package main



import (

	"bufio"

	"fmt"

	"github.com/spf13/cobra"

	"os"

)



const (

	Yellow = "\033[33m"

	Green  = "\033[32m"

	Reset  = "\033[0m"

)



var (

	token        string

	orgname      string

	output       string

	includeForks bool

	repoType     string

	member       string

	download     bool

	urlsOnly     bool

	usernamesOnly bool // new flag

	maxSizeMB    int

	provider     string // NEW: provider flag

	parallel      int // new flag for parallel cloning

)



func main() {

	var rootCmd = &cobra.Command{

		Use:   "github-org-tool",

		Short: "Fetch and report organization/group members and repositories from GitHub or GitLab.",

		Long: `GitHub/GitLab Organization Automation Tool

-----------------------------------

Fetch and report organization/group members and repositories from GitHub or GitLab with advanced filtering and output options.

Supports organization-wide and member-owned repositories, colored console output, and download link generation.



Main Features:

  - Fetch organization/group members and repositories (GitHub or GitLab)

  - Filter by member, repo type, and fork status

  - Print colored output and repo URLs

  - Download repositories (with size limit)

  - Output results to file

  - Read multiple orgs/groups from a file (pass filename to --orgname)

  - Print only repo URLs (--urls-only) or only usernames (--usernames-only)

  - Flexible repo type selection: org, member, both

  - All flags have short forms for usability

  - Select provider: --provider github|gitlab

`,

		Example: `

  # Fetch all GitHub org repos (default, forks excluded)

  github-org-tool --provider github --token <TOKEN> --orgname <ORG>



  # Fetch all GitLab group repos

  github-org-tool --provider gitlab --token <TOKEN> --orgname <GROUP>



  # Fetch all org/group repos, including forks

  github-org-tool --provider github --token <TOKEN> --orgname <ORG> --include-forks

  github-org-tool --provider gitlab --token <TOKEN> --orgname <GROUP> --include-forks



  # Fetch member-owned repos for all members/users

  github-org-tool --provider github --token <TOKEN> --orgname <ORG> --repo-type member

  github-org-tool --provider gitlab --token <TOKEN> --orgname <GROUP> --repo-type member



  # Print only repo URLs

  github-org-tool --provider github --token <TOKEN> --orgname <ORG> --urls-only

  github-org-tool --provider gitlab --token <TOKEN> --orgname <GROUP> --urls-only



  # Print only usernames of org/group members

  github-org-tool --provider github --token <TOKEN> --orgname <ORG> --usernames-only

  github-org-tool --provider gitlab --token <TOKEN> --orgname <GROUP> --usernames-only



  # Download all org/group repos (max size 250MB)

  github-org-tool --provider github --token <TOKEN> --orgname <ORG> --download

  github-org-tool --provider gitlab --token <TOKEN> --orgname <GROUP> --download



  # Save results to a file

  github-org-tool --provider github --token <TOKEN> --orgname <ORG> --output results.txt

  github-org-tool --provider gitlab --token <TOKEN> --orgname <GROUP> --output results.txt



  # Fetch for multiple orgs/groups listed in a file

  github-org-tool --provider github --token <TOKEN> --orgname orgs.txt --output results.txt

  github-org-tool --provider gitlab --token <TOKEN> --orgname groups.txt --output results.txt

`,

		Run: func(cmd *cobra.Command, args []string) {

			RunFetcher()

		},

	}



	rootCmd.Flags().StringVarP(&provider, "provider", "p", "github", "Provider to use: github or gitlab") // NEW

	rootCmd.Flags().StringVarP(&token, "token", "t", "", "Personal access token (optional for GitLab public info, required for GitHub/private)")

	rootCmd.Flags().StringVarP(&orgname, "orgname", "o", "", "Organization (GitHub) or group (GitLab) name (required)")

	rootCmd.Flags().StringVarP(&output, "output", "O", "", "Write results to output file")

	rootCmd.Flags().BoolVarP(&includeForks, "include-forks", "f", false, "Include forked repositories in the output")

	rootCmd.Flags().StringVarP(&repoType, "repo-type", "r", "org", "Type of repositories to fetch: 'org' for organization/group repos, 'member' for member/user-owned repos, 'both' for all")

	rootCmd.Flags().StringVarP(&member, "member", "m", "", "Username to fetch repos for particular user/member (only used with --repo-type member)")

	rootCmd.Flags().BoolVarP(&download, "download", "d", false, "Download all listed repositories using git clone")

	rootCmd.Flags().BoolVarP(&urlsOnly, "urls-only", "u", false, "Print only repo URLs (no names)")

	rootCmd.Flags().BoolVarP(&usernamesOnly, "usernames-only", "U", false, "Print only usernames of organization/group members (one per line)")

	rootCmd.Flags().IntVarP(&maxSizeMB, "max-size", "s", 250, "Maximum repo size (MB) to clone")

	rootCmd.Flags().IntVarP(&parallel, "parallel", "P", 4, "Number of concurrent clones when using --download")

	rootCmd.MarkFlagRequired("orgname")

	// Remove required flag for token if provider is gitlab

	cobra.OnInitialize(func() {

		if provider == "gitlab" && token == "" {

			// Don't require token for GitLab public info

			_ = rootCmd.Flags().SetAnnotation("token", cobra.BashCompOneRequiredFlag, nil)

		}

	})



	if err := rootCmd.Execute(); err != nil {

		fmt.Println(err)

		os.Exit(1)

	}

}



func getOrgList(orgname string) ([]string, error) {

	if fi, err := os.Stat(orgname); err == nil && !fi.IsDir() {

		// orgname is a file, read org names from file

		file, err := os.Open(orgname)

		if err != nil {

			return nil, err

		}

		defer file.Close()

		var orgs []string

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {

			line := scanner.Text()

			if line != "" {

				orgs = append(orgs, line)

			}

		}

		if err := scanner.Err(); err != nil {

			return nil, err

		}

		return orgs, nil

	}

	// orgname is a single org

	return []string{orgname}, nil

}



func RunFetcher() {

	orgs, err := getOrgList(orgname)

	if err != nil {

		fmt.Printf("Error reading orgname(s): %v\n", err)

		return

	}



	var f *os.File

	if output != "" {

		f, err = os.Create(output)

		if err != nil {

			fmt.Printf("Error creating output file: %v\n", err)

			return

		}

		defer f.Close()

	}



	totalMembers := 0

	totalRepos := 0

	totalMemberRepos := 0



	for _, org := range orgs {

		if provider == "gitlab" {

			// GitLab logic

			if usernamesOnly {

				members, err := fetchGitLabMembers(token, org)

				if err != nil {

					if f != nil {

						fmt.Fprintf(f, "Error fetching members: %v\n", err)

					} else {

						fmt.Printf("Error fetching members: %v\n", err)

					}

					continue

				}

				for _, m := range members {

					if f != nil {

						fmt.Fprintf(f, "%s\n", m.Username)

					} else {

						fmt.Println(m.Username)

					}

				}

				totalMembers += len(members)

				continue

			}

			if urlsOnly {

				if repoType == "org" {

					repos, err := fetchGitLabRepos(token, org)

					if err != nil {

						if f != nil {

							fmt.Fprintf(f, "Error fetching repos: %v\n", err)

						} else {

							fmt.Printf("Error fetching repos: %v\n", err)

						}

						continue

					}

					for _, r := range repos {

						if !includeForks && r.Fork {

							continue

						}

						if f != nil {

							fmt.Fprintf(f, "https://gitlab.com/%s/%s\n", org, r.Name)

						} else {

							fmt.Printf("https://gitlab.com/%s/%s\n", org, r.Name)

						}

						totalRepos++

					}

					continue

				} else if repoType == "member" {

					var members []string

					if member != "" {

						members = []string{member}

					} else {

						mlist, err := fetchGitLabMembers(token, org)

						if err != nil {

							if f != nil {

								fmt.Fprintf(f, "Error fetching members: %v\n", err)

							} else {

								fmt.Printf("Error fetching members: %v\n", err)

							}

							continue

						}

						for _, m := range mlist {

							members = append(members, m.Username)

						}

					}

					for _, m := range members {

						repos, err := fetchGitLabUserRepos(token, m)

						if err != nil {

							if f != nil {

								fmt.Fprintf(f, "Error fetching repos for %s: %v\n", m, err)

							} else {

								fmt.Printf("Error fetching repos for %s: %v\n", m, err)

							}

							continue

						}

						for _, r := range repos {

							if !includeForks && r.Fork {

								continue

							}

							if f != nil {

								fmt.Fprintf(f, "https://gitlab.com/%s/%s\n", m, r.Name)

							} else {

								fmt.Printf("https://gitlab.com/%s/%s\n", m, r.Name)

							}

							totalMemberRepos++

						}

					}

					continue

				} else if repoType == "both" {

					repos, err := fetchGitLabRepos(token, org)

					if err != nil {

						if f != nil {

							fmt.Fprintf(f, "Error fetching repos: %v\n", err)

						} else {

							fmt.Printf("Error fetching repos: %v\n", err)

						}

						continue

					}

					for _, r := range repos {

						if !includeForks && r.Fork {

							continue

						}

						if f != nil {

							fmt.Fprintf(f, "https://gitlab.com/%s/%s\n", org, r.Name)

						} else {

							fmt.Printf("https://gitlab.com/%s/%s\n", org, r.Name)

						}

						totalRepos++

					}

					var members []string

					if member != "" {

						members = []string{member}

					} else {

						mlist, err := fetchGitLabMembers(token, org)

						if err != nil {

							if f != nil {

								fmt.Fprintf(f, "Error fetching members: %v\n", err)

							} else {

								fmt.Printf("Error fetching members: %v\n", err)

							}

							continue

						}

						for _, m := range mlist {

							members = append(members, m.Username)

						}

					}

					for _, m := range members {

						repos, err := fetchGitLabUserRepos(token, m)

						if err != nil {

							if f != nil {

								fmt.Fprintf(f, "Error fetching repos for %s: %v\n", m, err)

							} else {

								fmt.Printf("Error fetching repos for %s: %v\n", m, err)

							}

							continue

						}

						for _, r := range repos {

							if !includeForks && r.Fork {

								continue

							}

							if f != nil {

								fmt.Fprintf(f, "https://gitlab.com/%s/%s\n", m, r.Name)

							} else {

								fmt.Printf("https://gitlab.com/%s/%s\n", m, r.Name)

							}

							totalMemberRepos++

						}

					}

					continue

				}

			}

			// ...other output modes for GitLab (full info, download, etc.)

			continue

		}

		if usernamesOnly {

			orgMembers, err := fetchMembers(token, org)

			if err != nil {

				if f != nil {

					fmt.Fprintf(f, "Error fetching members: %v\n", err)

				} else {

					fmt.Printf("Error fetching members: %v\n", err)

				}

				continue

			}

			for _, m := range orgMembers {

				if f != nil {

					fmt.Fprintf(f, "%s\n", m.Login)

				} else {

					fmt.Println(m.Login)

				}

			}

			totalMembers += len(orgMembers)

			continue

		}

		if urlsOnly {

			if repoType == "org" {

				repos, err := fetchRepos(token, org)

				if err != nil {

					if f != nil {

						fmt.Fprintf(f, "Error fetching repos: %v\n", err)

					} else {

						fmt.Printf("Error fetching repos: %v\n", err)

					}

					continue

				}

				for _, r := range repos {

					if !includeForks && r.Fork {

						continue

					}

					if f != nil {

						fmt.Fprintf(f, "https://github.com/%s/%s\n", org, r.Name)

					} else {

						fmt.Printf("https://github.com/%s/%s\n", org, r.Name)

					}

					totalRepos++

				}

				continue

			} else if repoType == "member" {

				var members []string

				if member != "" {

					members = []string{member}

				} else {

					orgMembers, err := fetchMembers(token, org)

					if err != nil {

						if f != nil {

							fmt.Fprintf(f, "Error fetching members: %v\n", err)

						} else {

							fmt.Printf("Error fetching members: %v\n", err)

						}

						continue

					}

					for _, m := range orgMembers {

						members = append(members, m.Login)

					}

				}

				for _, m := range members {

					repos, err := fetchUserRepos(token, m)

					if err != nil {

						if f != nil {

							fmt.Fprintf(f, "Error fetching repos for %s: %v\n", m, err)

						} else {

							fmt.Printf("Error fetching repos for %s: %v\n", m, err)

						}

						continue

					}

					for _, r := range repos {

						if !includeForks && r.Fork {

							continue

						}

						if f != nil {

							fmt.Fprintf(f, "https://github.com/%s/%s\n", m, r.Name)

						} else {

							fmt.Printf("https://github.com/%s/%s\n", m, r.Name)

						}

						totalMemberRepos++

					}

				}

				continue

			} else if repoType == "both" {

				repos, err := fetchRepos(token, org)

				if err != nil {

					if f != nil {

						fmt.Fprintf(f, "Error fetching org repos: %v\n", err)

					} else {

						fmt.Printf("Error fetching org repos: %v\n", err)

					}

				} else {

					for _, r := range repos {

						if !includeForks && r.Fork {

							continue

						}

						if f != nil {

							fmt.Fprintf(f, "https://github.com/%s/%s\n", org, r.Name)

						} else {

							fmt.Printf("https://github.com/%s/%s\n", org, r.Name)

						}

						totalRepos++

					}

				}

				var members []string

				if member != "" {

					members = []string{member}

				} else {

					orgMembers, err := fetchMembers(token, org)

					if err != nil {

						if f != nil {

							fmt.Fprintf(f, "Error fetching members: %v\n", err)

						} else {

							fmt.Printf("Error fetching members: %v\n", err)

						}

						continue

					}

					for _, m := range orgMembers {

						members = append(members, m.Login)

					}

				}

				for _, m := range members {

					repos, err := fetchUserRepos(token, m)

					if err != nil {

						if f != nil {

							fmt.Fprintf(f, "Error fetching repos for %s: %v\n", m, err)

						} else {

							fmt.Printf("Error fetching repos for %s: %v\n", m, err)

						}

						continue

					}

					for _, r := range repos {

						if !includeForks && r.Fork {

							continue

						}

						if f != nil {

							fmt.Fprintf(f, "https://github.com/%s/%s\n", m, r.Name)

						} else {

							fmt.Printf("https://github.com/%s/%s\n", m, r.Name)

						}

						totalMemberRepos++

					}

				}

				continue

			}

		}

		if repoType == "org" {

			repos, err := fetchRepos(token, org)

			if err != nil {

				if f != nil {

					fmt.Fprintf(f, "Error fetching repos: %v\n", err)

				} else {

					fmt.Printf("Error fetching repos: %v\n", err)

				}

				continue

			}

			if f != nil {

				fmt.Fprintf(f, "Organization: %s\n", org)

			} else {

				fmt.Printf("Organization: %s\n", org)

			}

			for _, r := range repos {

				if !includeForks && r.Fork {

					continue

				}

				if f != nil {

					fmt.Fprintf(f, "Repo: %s\n", r.Name)

					fmt.Fprintf(f, "  URL: https://github.com/%s/%s\n", org, r.Name)

					fmt.Fprintf(f, "  Fork: %v\n", r.Fork)

					fmt.Fprintf(f, "  Size (KB): %d\n", r.Size)

					fmt.Fprintf(f, "  Owner: %s\n", r.Owner.Login)

				} else {

					fmt.Printf("Repo: %s\n", r.Name)

					fmt.Printf("  URL: https://github.com/%s/%s\n", org, r.Name)

					fmt.Printf("  Fork: %v\n", r.Fork)

					fmt.Printf("  Size (KB): %d\n", r.Size)

					fmt.Printf("  Owner: %s\n", r.Owner.Login)

				}

			}

			members, err := fetchMembers(token, org)

			if err != nil {

				if f != nil {

					fmt.Fprintf(f, "Error fetching members: %v\n", err)

				} else {

					fmt.Printf("Error fetching members: %v\n", err)

				}

				continue

			}

			for _, m := range members {

				if f != nil {

					fmt.Fprintf(f, "Member: %s\n", m.Login)

				} else {

					fmt.Printf("Member: %s\n", m.Login)

				}

			}

			totalMembers += len(members)

			continue

		} else if repoType == "member" {

			var members []string

			if member != "" {

				members = []string{member}

			} else {

				orgMembers, err := fetchMembers(token, org)

				if err != nil {

					if f != nil {

						fmt.Fprintf(f, "Error fetching members: %v\n", err)

					} else {

						fmt.Printf("Error fetching members: %v\n", err)

					}

					continue

				}

				for _, m := range orgMembers {

					members = append(members, m.Login)

				}

			}

			for _, m := range members {

				repos, err := fetchUserRepos(token, m)

				if err != nil {

					if f != nil {

						fmt.Fprintf(f, "Error fetching repos for %s: %v\n", m, err)

					} else {

						fmt.Printf("Error fetching repos for %s: %v\n", m, err)

					}

					continue

				}

				for _, r := range repos {

					if !includeForks && r.Fork {

						continue

					}

					if f != nil {

						fmt.Fprintf(f, "%s/%s\n", m, r.Name)

					} else {

						fmt.Printf("%s/%s\n", m, r.Name)

					}

					totalMemberRepos++

				}

			}

			for _, m := range members {

				if f != nil {

					fmt.Fprintf(f, "Member: %s\n", m)

				} else {

					fmt.Printf("Member: %s\n", m)

				}

			}

			continue

		} else if repoType == "both" {

			repos, err := fetchRepos(token, org)

			if err != nil {

				if f != nil {

					fmt.Fprintf(f, "Error fetching org repos: %v\n", err)

				} else {

					fmt.Printf("Error fetching org repos: %v\n", err)

				}

			} else {

				if f != nil {

					fmt.Fprintf(f, "Organization: %s\n", org)

				} else {

					fmt.Printf("Organization: %s\n", org)

				}

				for _, r := range repos {

					if !includeForks && r.Fork {

						continue

					}

					if f != nil {

						fmt.Fprintf(f, "Repo: %s\n", r.Name)

						fmt.Fprintf(f, "  URL: https://github.com/%s/%s\n", org, r.Name)

						fmt.Fprintf(f, "  Fork: %v\n", r.Fork)

						fmt.Fprintf(f, "  Size (KB): %d\n", r.Size)

						fmt.Fprintf(f, "  Owner: %s\n", r.Owner.Login)

					} else {

						fmt.Printf("Repo: %s\n", r.Name)

						fmt.Printf("  URL: https://github.com/%s/%s\n", org, r.Name)

						fmt.Printf("  Fork: %v\n", r.Fork)

						fmt.Printf("  Size (KB): %d\n", r.Size)

						fmt.Printf("  Owner: %s\n", r.Owner.Login)

					}

				}

			}

			var members []string

			if member != "" {

				members = []string{member}

			} else {

				orgMembers, err := fetchMembers(token, org)

				if err != nil {

					if f != nil {

						fmt.Fprintf(f, "Error fetching members: %v\n", err)

					} else {

						fmt.Printf("Error fetching members: %v\n", err)

					}

					continue

				}

				for _, m := range orgMembers {

					members = append(members, m.Login)

				}

			}

			for _, m := range members {

				repos, err := fetchUserRepos(token, m)

				if err != nil {

					if f != nil {

						fmt.Fprintf(f, "Error fetching repos for %s: %v\n", m, err)

					} else {

						fmt.Printf("Error fetching repos for %s: %v\n", m, err)

					}

					continue

				}

				for _, r := range repos {

					if !includeForks && r.Fork {

						continue

					}

					if f != nil {

						fmt.Fprintf(f, "%s/%s\n", m, r.Name)

					} else {

						fmt.Printf("%s/%s\n", m, r.Name)

					}

					totalMemberRepos++

				}

			}

		}

	}

	// Print totals to console only (not to file)

	if usernamesOnly {

		fmt.Printf("%sTotal members: %s%d%s\n", Yellow, Green, totalMembers, Reset)

	} else if urlsOnly {

		if repoType == "org" || repoType == "both" {

			fmt.Printf("%sTotal repositories: %s%d%s\n", Yellow, Green, totalRepos, Reset)

		}

		if repoType == "member" || repoType == "both" {

			fmt.Printf("%sTotal member-owned repositories: %s%d%s\n", Yellow, Green, totalMemberRepos, Reset)

		}

	}

	// Parallel download logic

	if download {

		var repoURLs []string

		mainFolder := "downloaded_repos"

		_ = os.MkdirAll(mainFolder, 0755)

		// Collect all repo URLs to clone

		for _, org := range orgs {

			if provider == "github" {

				repos, err := fetchRepos(token, org)

				if err != nil {

					fmt.Printf("Error fetching repos: %v\n", err)

					continue

				}

				for _, r := range repos {

					if !includeForks && r.Fork {

						continue

					}

					url := fmt.Sprintf("https://github.com/%s/%s", org, r.Name)

					repoURLs = append(repoURLs, url)

				}

				if repoType == "member" || repoType == "both" {

					var members []string

					if member != "" {

						members = []string{member}

					} else {

						orgMembers, err := fetchMembers(token, org)

						if err != nil {

							fmt.Printf("Error fetching members: %v\n", err)

							continue

						}

						for _, m := range orgMembers {

							members = append(members, m.Login)

						}

					}

					for _, m := range members {

						repos, err := fetchUserRepos(token, m)

						if err != nil {

							fmt.Printf("Error fetching repos for %s: %v\n", m, err)

							continue

						}

						for _, r := range repos {

							if !includeForks && r.Fork {

								continue

							}

							url := fmt.Sprintf("https://github.com/%s/%s", m, r.Name)

							repoURLs = append(repoURLs, url)

						}

					}

				}

			} else if provider == "gitlab" {

				repos, err := fetchGitLabRepos(token, org)

				if err != nil {

					fmt.Printf("Error fetching repos: %v\n", err)

					continue

				}

				for _, r := range repos {

					if !includeForks && r.Fork {

						continue

					}

					url := fmt.Sprintf("https://gitlab.com/%s/%s", org, r.Name)

					repoURLs = append(repoURLs, url)

				}

				if repoType == "member" || repoType == "both" {

					var members []string

					if member != "" {

						members = []string{member}

					} else {

						mlist, err := fetchGitLabMembers(token, org)

						if err != nil {

							fmt.Printf("Error fetching members: %v\n", err)

							continue

						}

						for _, m := range mlist {

							members = append(members, m.Username)

						}

					}

					for _, m := range members {

						repos, err := fetchGitLabUserRepos(token, m)

						if err != nil {

							fmt.Printf("Error fetching repos for %s: %v\n", m, err)

							continue

						}

						for _, r := range repos {

							if !includeForks && r.Fork {

								continue

							}

							url := fmt.Sprintf("https://gitlab.com/%s/%s", m, r.Name)

							repoURLs = append(repoURLs, url)

						}

					}

				}

			}

		}

		// Parallel worker pool

		type job struct {

			url string

		}

		jobs := make(chan job, len(repoURLs))

		results := make(chan string, len(repoURLs))

		for w := 0; w < parallel; w++ {

			go func() {

				for j := range jobs {

					err := cloneRepo(j.url, mainFolder)

					if err != nil {

						results <- fmt.Sprintf("Failed: %s (%v)", j.url, err)

					} else {

						results <- fmt.Sprintf("Cloned: %s", j.url)

					}

				}

			}()

		}

		for _, url := range repoURLs {

			jobs <- job{url: url}

		}

		close(jobs)

		for i := 0; i < len(repoURLs); i++ {

			fmt.Println(<-results)

		}

		return // skip rest of RunFetcher if download is used

	}

}

