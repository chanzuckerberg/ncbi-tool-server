package models

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"ncbi-tool-server/utils"
	"path"
	"sort"
	"strings"
)

// Directory Model
type Directory struct {
	ctx *utils.Context
}

// NewDirectory makes a new directory instance
func NewDirectory(ctx *utils.Context) *Directory {
	return &Directory{
		ctx: ctx,
	}
}

// GetPast gets the approximate directory listing at a point in time
// from the Db.
func (d *Directory) GetPast(pathName string, inputTime string,
	output string) ([]Entry, error) {
	// Setup
	var err error
	resp := []Entry{}
	file := NewFile(d.ctx)
	url := ""

	// Get archive versions from DB
	listing, err := d.getAtTimeDb(pathName, inputTime)
	if err != nil || len(listing) == 0 {
		return resp, errors.New("empty or non-existent directory")
	}

	// Process results
	downloadName := path.Base(pathName)
	for _, val := range listing {
		if output == "with-URLs" {
			key := file.getS3Key(val)
			url, err = file.keyToURL(key, downloadName)
			if err != nil {
				return resp, err
			}
		}
		entry := Entry{
			Path:    val.Path,
			Version: val.Version,
			ModTime: val.ModTime.String,
			URL:     url,
		}
		resp = append(resp, entry)
	}

	if len(resp) == 0 {
		err = errors.New("no results")
	}
	return resp, err
}

// Gets the approximate directory state at a given time. Finds the
// most recent version of each file in a path before a given date.
func (d *Directory) getAtTimeDb(pathName string,
	inputTime string) ([]Metadata, error) {
	// Query
	res := []Metadata{}
	rows, err := d.ctx.Db.Query("select e.PathName, e.VersionNum, "+
		"e.DateModified, e.ArchiveKey "+
		"from entries as e "+
		"inner join ( "+
		"select max(VersionNum) VersionNum, PathName "+
		"from entries "+
		"where PathName LIKE ? "+
		"and DateModified <= ? "+
		"group by PathName ) as max "+
		"on max.PathName = e.PathName "+
		"and max.VersionNum = e.VersionNum",
		pathName+"%%", inputTime)
	if err != nil {
		return res, errors.New("no results found")
	}
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			err = utils.ComboErr("Couldn't close db rows.", closeErr, err)
			log.Println(err)
		}
	}()

	// Process results
	for rows.Next() {
		md := Metadata{}
		err = rows.Scan(&md.Path, &md.Version, &md.ModTime, &md.ArchiveKey)
		if err != nil {
			return res, err
		}
		// Exclude files in sub-folders with another folder slash
		rest := md.Path[len(pathName):]
		if strings.Contains(rest, "/") {
			continue
		}
		res = append(res, md)
	}
	return res, err
}

// CompareResponse is for comparing directory diffs between times. Tag is
// used for labeling files as added, updated, or unchanged.
type CompareResponse struct {
	Path string
	Tag  string
}

// CompareListing compares the directory state betwen the startDate and
// endDate and returns a file listing of CompareResponses.
func (d *Directory) CompareListing(pathName string, startDate string,
	endDate string) ([]CompareResponse, error) {
	result := []CompareResponse{}

	// Get approximate file listings at start and end dates
	// Get a mapping of file name -> version num. Default is
	// zero value.
	startSet, err := d.getListingAtTime(pathName, startDate)
	if err != nil {
		err = utils.NewErr("Error in getting listing at time.", err)
		log.Print(err)
		return result, err
	}
	endSet, err := d.getListingAtTime(pathName, endDate)
	if err != nil {
		err = utils.NewErr("Error in getting listing at time.", err)
		return result, err
	}

	// Compare file-by-file. Sort keys to be in order.
	keys := []string{}
	for k := range endSet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, file := range keys {
		if startSet[file] == 0 {
			// Present in endSet but not startSet means added since start date
			result = append(result, CompareResponse{file, "Added"})
		} else {
			// Present in both sets
			if startSet[file] == endSet[file] {
				// Same VersionNum, so unchanged
				result = append(result, CompareResponse{file, "Unchanged"})
			} else {
				// Different VersionNum, so modified
				result = append(result, CompareResponse{file, "Updated"})
			}
		}
	}
	return result, err
}

// Get a list of files present at a time in a directory
func (d *Directory) getListingAtTime(pathName string,
	inputTime string) (map[string]int, error) {
	// Query
	listing := make(map[string]int)
	rows, err := d.ctx.Db.Query("select e.PathName, e.VersionNum "+
		"from entries as e "+
		"inner join ( "+
		"select max(VersionNum) VersionNum, PathName "+
		"from entries "+
		"where PathName LIKE ? "+
		"and DateModified <= ? "+
		"group by PathName ) as max "+
		"on max.PathName = e.PathName "+
		"and max.VersionNum = e.VersionNum",
		pathName+"%%", inputTime)
	if err != nil {
		return listing, utils.NewErr("No results found at time "+inputTime+".", err)
	}
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			err = utils.ComboErr("Couldn't close db rows.", closeErr, err)
			log.Println(err)
		}
	}()

	// Process results
	for rows.Next() {
		md := Metadata{}
		err = rows.Scan(&md.Path, &md.Version)
		if err != nil {
			err = utils.NewErr("Error in scanning rows.", err)
			return listing, err
		}
		listing[md.Path] = md.Version
	}
	return listing, err
}
