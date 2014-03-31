package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"
)

var aurURL, _ = url.Parse("https://aur.archlinux.org")

type Package struct {
	Maintainer     string
	ID             int
	Name           string
	Version        string
	CategoryID     int
	Description    string
	URL            string
	License        string
	NumVotes       int
	OutOfDate      int // actually a boolean but the JSON response is 0/1
	FirstSubmitted timeUnmarshaler
	LastModified   timeUnmarshaler
	URLPath        string
}

type PackageList []Package

func (p PackageList) Len() int      { return len(p) }
func (p PackageList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// implements the same sorting order as https://aur.archlinux.org/packages/
func (p PackageList) Less(i, j int) bool {
	if p[i].CategoryID < p[j].CategoryID {
		// sort by category
		return true
	}
	if p[i].CategoryID == p[j].CategoryID && p[i].Name < p[j].Name {
		// items in same category get sorted by name
		return true
	}
	return false
}

type Response struct {
	Type    string
	Count   int
	Results []Package
}

// unmarshals time.Time in the Unix format instead of RFC3339
type timeUnmarshaler struct{ time.Time }

func (t *timeUnmarshaler) UnmarshalJSON(str []byte) error {
	unix, err := strconv.Atoi(string(str))
	if err != nil {
		return err
	}

	t.Time = time.Unix(int64(unix), 0)
	return nil
}

// aurRequest takes a string representing an RPC request to the AUR, and
// returns a Response containing the results of the request. The string must
// already be escaped where desired, eg "/rpc.php?type=search&arg=jquery".
func aurRequest(request string) (r Response, err error) {
	requestURL, err := url.Parse(request)
	if err != nil {
		return
	}

	resp, err := http.Get(aurURL.ResolveReference(requestURL).String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&r)
	return
}

// multiInfo takes a slice of strings, each string exactly matching the name of
// an AUR package, and a function to perform on Packages. It performs a
// multiinfo request to the AUR to retrieve information on the packages listed,
// and invokes f on each one, displaying an error for any packages that do not
// exist.
// If the request was successful and all the packages existed, multiInfo will
// return nil. If one or more of the listed packages did not exist, multiInfo
// returns the error "Some packages were not found".
func multiInfo(args []string, f func(Package) error) error {
	request := "/rpc.php?type=multiinfo"
	for _, str := range args {
		request += "&arg[]="
		request += url.QueryEscape(str)
	}

	results, err := aurRequest(request)
	if err != nil {
		return err
	}

	// order of packages in the response is arbitrary - it might not match
	// the order of the arguments
	// so you have to search for each argument through the entire response
	for arg := 0; arg < len(args); arg++ {
		found := false
		for _, pkg := range results.Results {
			if pkg.Name == args[arg] {
				err := f(pkg)
				if err != nil {
					return err
				}
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("\x1B[1;31merror:\x1B[0m package '%s' was not found\n", args[arg])
		}
	}

	if len(results.Results) < len(args) {
		return fmt.Errorf("Some packages were not found")
	}
	return nil
}

func main() {
	categories := []string{
		1:  "none",
		2:  "daemons",
		3:  "devel",
		4:  "editors",
		5:  "emulators",
		6:  "games",
		7:  "gnome",
		8:  "18n",
		9:  "kde",
		10: "lib",
		11: "modules",
		12: "multimedia",
		13: "network",
		14: "office",
		15: "science",
		16: "system",
		17: "x11",
		18: "xfce",
		19: "kernels",
		20: "fonts",
	}

	var search = &cobra.Command{
		Use:   "search [string to search]",
		Short: "Search for packages whose name contains the argument",
		Long:  `Displays the list of packages whose names contain the argument.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				println("search must be invoked with exactly one argument")
				os.Exit(1)
			}

			results, err := aurRequest("/rpc.php?type=search&arg=" + url.QueryEscape(args[0]))
			if err != nil {
				panic(err)
			}

			sort.Sort(PackageList(results.Results))
			for _, p := range results.Results {
				fmt.Printf("%saur/%s/%s%s ", "\x1B[1;35m", categories[p.CategoryID], "\x1B[1;37m", p.Name)

				if p.OutOfDate == 0 {
					// the package is up to date
					fmt.Print("\x1B[1;32m") // green
				} else {
					fmt.Print("\x1B[1;31m") // red
				}
				fmt.Println(p.Version)

				fmt.Printf("    %s%s\n", "\x1B[0m", p.Description)
			}
		},
	}

	var info = &cobra.Command{
		Use:   "info [names of packages]",
		Short: "Retrieve info for the given packages",
		Long: `Displays detailed information for each package specified. Each argument must
exactly match a package name. Arguments with no corresponding package will
cause an error to be displayed.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				println("info requires at least one argument")
				os.Exit(1)
			}

			err := multiInfo(args, func(thispkg Package) error {
				fmt.Printf("\x1B[1mCategory        : \x1B[0m%s\n", categories[thispkg.CategoryID])
				fmt.Printf("\x1B[1mName            : \x1B[0m%s\n", thispkg.Name)
				fmt.Printf("\x1B[1mVersion         : \x1B[0m%s", thispkg.Version)
				if thispkg.OutOfDate != 0 {
					fmt.Print(" [out of date]")
				}
				fmt.Println()
				fmt.Printf("\x1B[1mDescription     : \x1B[0m%s\n", thispkg.Description)
				fmt.Printf("\x1B[1mURL             : \x1B[0m%s\n", thispkg.URL)
				fmt.Printf("\x1B[1mLicenses        : \x1B[0m%s\n", thispkg.License)
				fmt.Printf("\x1B[1mMaintainer      : \x1B[0m%s\n", thispkg.Maintainer)
				fmt.Printf("\x1B[1mFirst Submitted : \x1B[0m%s\n", thispkg.FirstSubmitted.Format("Mon 02 Jan 2006 03:04:05 PM MST"))
				fmt.Printf("\x1B[1mLast Modified   : \x1B[0m%s\n", thispkg.LastModified.Format("Mon 02 Jan 2006 03:04:05 PM MST"))
				fmt.Printf("\x1B[1mVotes           : \x1B[0m%v\n", thispkg.NumVotes)
				fmt.Println()
				return nil
			})

			if err != nil {
				os.Exit(1)
			}
		},
	}

	var tarball = &cobra.Command{
		Use:   "tarball [names of packages]",
		Short: "Retrieve tarball link for the given packages",
		Long: `Display the tarball link for the given packages. Each argument must exactly
match a package name. Arguments with no corresponding package will cause an
error to be displayed.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				println("tarball requires at least one argument")
				os.Exit(1)
			}

			//err := multiInfo(args, func(thispkg Package) error {
			//	fmt.Printf("https://aur.archlinux.org%s\n", thispkg.URLPath)
			//	return nil
			//})
			//if err != nil {
			//	os.Exit(1)
			//}

			for _, s := range args {
				fmt.Printf("https://aur.archlinux.org/packages/%s/%s/%s.tar.gz\n", s[0:2], s, s)
			}
		},
	}

	var pkgbuild = &cobra.Command{
		Use:   "pkgbuild [names of packages]",
		Short: "Display PKGBUILDs for the given packages",
		Long: `Display the PKGBUILDs for the given packages. Each argument must exactly match
a package name. Arguments with no corresponding package will cause an error to
be displayed.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				println("pkgbuild requires at least one argument")
				os.Exit(1)
			}

			printPKGBUILD := func(s string) error {
				resp, err := http.Get(s)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				pkgb, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				fmt.Println(string(pkgb))
				return nil
			}

			//err := multiInfo(args, func(thispkg Package) error {
			//	aurURL, _ := url.Parse("https://aur.archlinux.org")
			//	requestURL, err := url.Parse(thispkg.URLPath + "/../PKGBUILD")
			//	if err != nil {
			//		return err
			//	}
			//	return printPKGBUILD(aurURL.ResolveReference(requestURL).String())
			//})
			//if err != nil {
			//	os.Exit(1)
			//}

			for _, s := range args {
				err := printPKGBUILD("https://aur.archlinux.org/packages/" + s[0:2] + "/" + s + "/PKGBUILD")
				if err != nil {
					os.Exit(1)
				}
			}
		},
	}

	var version = &cobra.Command{
		Use:   "version",
		Short: "Displays the version",
		Long:  "Displays the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("seekaur v1.0.0")
		},
	}

	var root = &cobra.Command{Use: "seekaur"}
	root.AddCommand(search, info, tarball, pkgbuild, version)
	root.Execute()
}
