package ghmetrics

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	reposRegex          = regexp.MustCompile(`^/repos/(?P<owner>[^/]*)/(?P<repo>[^/]*)(?P<rest>/.*)$`)
	varAndConstantRegex = regexp.MustCompile(`^/(?P<var>[^/]*)(?P<path>/.*)?$`)
	constantAndVarRegex = regexp.MustCompile(`^/(?P<path>[^/]*)(?P<var>/.*)?$`)
)

func getSimplifiedPath(path string) string {
	fragment := getFirstFragment(path)
	switch fragment {
	case "repos":
		// /:owner/:repo
		return handleRepos(path)
	case "user":
		// do some more
		return path
	case "users":
		// do some more
		return path
	case "orgs":
		// do some more
		return path
	case "issues":
		// do some more
		return path
	case "search":
		// do some more
		return path
	case "gists":
		// do some more
		return path
	case "notifications":
		// do some more
		return path
	case "repositories", "emojis", "events", "feeds", "hub", "rate_limits", "teams", "licenses":
		return path
	default:
		logrus.WithField("path", path).Warning("Path not handled")
		return path
	}
}

// getFirstFragment returns the first fragment of a path, of a
// `*url.URL.Path`. e.g. `/repos/kubernetes/test-infra/` will return `repos`
func getFirstFragment(path string) string {
	re := regexp.MustCompile(`^/(?P<path>[^/]*).*$`)
	fragment := re.FindStringSubmatch(path)
	if len(fragment) < 2 {
		return path
	}
	return fragment[1]
}

func handleRepos(path string) string {
	match := reposRegex.FindStringSubmatch(path)
	result := make(map[string]string)
	for i, name := range reposRegex.SubexpNames() {
		if i != 0 && name != "" && i <= len(match) { // skip first and empty
			if match[i] != "" {
				result[name] = match[i]
			}
		}
	}
	if result["owner"] == "" || result["repo"] == "" {
		logrus.WithField("path", path).Warning("Not handling /repos/.. path correctly")
		return "/repos"
	}
	rest := result["rest"]
	sanitizedPath := fmt.Sprintf("/repos/%s/%s", ":owner", ":repo")
	if rest == "" || rest == "/" {
		return sanitizedPath
	}
	switch fragment := getFirstFragment(rest); fragment {
	case "issues", "branches":
		return fmt.Sprintf("%s%s", sanitizedPath, handlePrefixedVarAndConstant(fmt.Sprintf("/%s", fragment), rest))
	case "keys", "labels", "milestones", "pulls", "releases", "statuses", "subscribers", "assignees", "archive":
		// archive is a special path that might need better handling
		return fmt.Sprintf("%s/%s%s", sanitizedPath, fragment, handleConstantAndVar(rest))
	case "git":
		return fmt.Sprintf("%s%s", sanitizedPath, handlePrefixedConstantAndVar(fmt.Sprintf("/%s", fragment), rest))

	case "merges", "stargazers", "notifications", "hooks":
		return fmt.Sprintf("%s%s", sanitizedPath, rest)
	default:
		logrus.WithField("sanitizedPath", sanitizedPath).WithField("rest", rest).Warning("Path not handled")
		return fmt.Sprintf("%s%s", sanitizedPath, rest)
	}
}

func handlePrefixedVarAndConstant(prefix, path string) string {
	pathTrimmed := strings.Replace(path, prefix, "", 1)
	fragment := getFirstFragment(pathTrimmed)
	switch fragment {
	case "comments", "events":
		return fmt.Sprintf("%s/%s%s", prefix, fragment, handleConstantAndVar(pathTrimmed))
	}

	match := varAndConstantRegex.FindStringSubmatch(pathTrimmed)
	result := make(map[string]string)
	for i, name := range varAndConstantRegex.SubexpNames() {
		if i != 0 && name != "" && i <= len(match) { // skip first and empty
			if name == "var" && match[i] != "" {
				result[name] = ":var" // mask issue number
			} else {
				result[name] = match[i]
			}
		}
	}
	rest := result["path"]
	sanitizedPath := fmt.Sprintf("%s/%s", prefix, result["var"])
	if result["var"] == "" && rest == "" {
		return prefix
	} else if rest == "" || rest == "/" {
		return sanitizedPath
	}
	return fmt.Sprintf("%s/%s%s", prefix, result["var"], rest)
}

func handlePrefixedConstantAndVar(prefix, path string) string {
	pathTrimmed := strings.Replace(path, prefix, "", 1)
	fragment := getFirstFragment(pathTrimmed)
	if fragment != "" {
		return fmt.Sprintf("%s/%s%s", prefix, fragment, handleConstantAndVar(pathTrimmed))
	}

	match := varAndConstantRegex.FindStringSubmatch(pathTrimmed)
	result := make(map[string]string)
	for i, name := range varAndConstantRegex.SubexpNames() {
		if i != 0 && name != "" && i <= len(match) { // skip first and empty
			if name == "var" && match[i] != "" {
				result[name] = ":var" // mask issue number
			} else {
				result[name] = match[i]
			}
		}
	}
	rest := result["path"]
	sanitizedPath := fmt.Sprintf("%s/%s", prefix, result["var"])
	if result["var"] == "" && rest == "" {
		return prefix
	} else if rest == "" || rest == "/" {
		return sanitizedPath
	}
	return fmt.Sprintf("%s/%s%s", prefix, result["var"], rest)
}

func handleConstantAndVar(path string) string {
	match := constantAndVarRegex.FindStringSubmatch(path)
	result := make(map[string]string)
	for i, name := range constantAndVarRegex.SubexpNames() {
		if i != 0 && name != "" && i <= len(match) { // skip first and empty
			if name == "var" && match[i] != "" {
				result[name] = ":var" // mask issue number
			} else {
				result[name] = match[i]
			}
		}
	}
	if result["var"] != "" {
		return fmt.Sprintf("/%s", result["var"])
	}
	return ""
}

func handleVarAndConstant(path string) string {
	match := varAndConstantRegex.FindStringSubmatch(path)
	result := make(map[string]string)
	for i, name := range varAndConstantRegex.SubexpNames() {
		if i != 0 && name != "" && i <= len(match) { // skip first and empty
			if name == "var" && match[i] != "" {
				result[name] = ":var" // mask issue number
			} else {
				result[name] = match[i]
			}
		}
	}
	if result["var"] != "" {
		return fmt.Sprintf("/%s", result["var"])
	}
	return result["constant"]
}