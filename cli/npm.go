package cli

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/levigross/grequests"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func (n *Nexus3) BackupAllNPMArtifacts(repository, dir string) error {
	url := n.URL + "/service/rest/repository/browse/" + repository
	npmRepoHTML, err := n.npmURL(url)
	if err != nil {
		return err
	}
	log.Debugf("NPM URL HTML: '%v'", npmRepoHTML)

	npmArtifactDirectoriesHTMLNodes, err := npmArtifactRepositories(npmRepoHTML)
	if err != nil {
		return err
	}
	log.Debugf("npmArtifactDirectoriesHTMLNodes: '%v'", npmArtifactDirectoriesHTMLNodes)

	n.boo(npmArtifactDirectoriesHTMLNodes, url, dir)

	return nil
}

func (n *Nexus3) npmURL(url string) (string, error) {
	resp, err := grequests.Get(url, &grequests.RequestOptions{Auth: []string{n.User, n.Pass}})
	if err != nil {
		return "", err
	}

	statusCode := resp.StatusCode
	log.Debugf("URL: '%v'. StatusCode: '%v'", url, statusCode)
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("StatusCode URL: '%s' not OK, but: '%d'", url, statusCode)
	}
	return resp.String(), nil
}

func npmArtifactRepositories(s string) ([]*html.Node, error) {
	r := strings.NewReader(s)
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	bodies := cascadia.MustCompile("tr td a").MatchAll(doc)
	log.Debugf("npmArtifactRepositories: '%v'", bodies)
	return bodies, nil
}

func (n *Nexus3) boo(npmArtifactDirectoriesHTMLNodes []*html.Node, url, dir string) {
	errs := make(chan error)
	for _, npmArtifactDirectoriesHTMLNode := range npmArtifactDirectoriesHTMLNodes {
		npmArtifactDirectory := goquery.NewDocumentFromNode(npmArtifactDirectoriesHTMLNode).Text()
		log.Debugf("npmArtifactDirectory: '%v'", npmArtifactDirectory)
		n.wat(errs, npmArtifactDirectory, url, dir)
	}
}

//
//
//
//
//
func (n *Nexus3) Bladibla(url, dir string) error {
	resp, err := grequests.Get(url, &grequests.RequestOptions{Auth: []string{n.User, n.Pass}})
	if err != nil {
		return err
	}

	statusCode := resp.StatusCode
	if statusCode != http.StatusOK {
		log.Errorf("StatusCode URL: '%s' not OK, but: '%d'", url, statusCode)
	}

	r := strings.NewReader(resp.String())
	doc, err := html.Parse(r)
	if err != nil {
		return err
	}

	bodies := cascadia.MustCompile("tr td a").MatchAll(doc)
	for _, body := range bodies {
		errs := make(chan error)
		go func(n *Nexus3, body *html.Node, url string) {
			log.Debugf("Go Channel length (inside go routine): '%d'", len(errs))
			n.wat(errs, "", url, dir)
		}(n, body, url)
		time.Sleep(100 * time.Millisecond)
		d := <-errs
		log.Debug("Main goroutine received data:", d)
	}

	// if err := <-errs; err != nil {
	// 	return err
	// }
	return nil
}

func (n *Nexus3) wat(errs chan error, s, url, dir string) error {
	if s != "Parent Directory" {
		log.Debug(s)
		url2 := url + "/" + s
		log.Debug("URL: ", url2)
		log.Debug("Extension: ", filepath.Ext(url2))

		if filepath.Ext(url2) == ".tgz" {
			go func(errs chan error) {
				re, err := regexp.Compile("^(.*)/service\\/rest\\/repository\\/browse\\/(.*)\\/(.*)$")
				if err != nil {
					// return err
					log.Error("===============================")
				}
				if !re.MatchString(url2) {
					// return fmt.Errorf("No MATCH!!!!!!!!!!: %v", url2)
					log.Error("===============================")
				}
				group := re.FindStringSubmatch(url2)
				url2 = group[1] + "/repository/" + group[2] + "/-/" + group[3]

				log.Debug("Download URL: " + url2)
				resp, err := grequests.Get(url2, &grequests.RequestOptions{Auth: []string{n.User, n.Pass}})
				if err != nil {
					// return err
					log.Error("===============================")
				}
				log.Debug("FILEPATH", filepath.Join(dir, group[2], group[3]))
				os.MkdirAll(filepath.Join(dir, group[2]), os.ModePerm)
				if err := resp.DownloadToFile(filepath.Join(dir, group[2], group[3])); err != nil {
					// return err
					log.Error("===============================")
				}
				fmt.Print("+")
				errs <- nil
			}(errs)
			if err := <-errs; err != nil {
				return err
			}
		}

		aaa, err := n.npmURL(url2)
		if err != nil {
			return err
		}
		npmArtifactDirectoriesHTMLNodes, err := npmArtifactRepositories(aaa)
		if err != nil {
			return err
		}
		log.Debugf("npmArtifactDirectoriesHTMLNodes: '%v'", npmArtifactDirectoriesHTMLNodes)

		for _, npmArtifactDirectoriesHTMLNode := range npmArtifactDirectoriesHTMLNodes {
			npmArtifactDirectory := goquery.NewDocumentFromNode(npmArtifactDirectoriesHTMLNode).Text()
			log.Debugf("npmArtifactDirectory: '%v'", npmArtifactDirectory)
			n.wat(errs, npmArtifactDirectory, url2, dir)
		}
	}

	return nil
}
