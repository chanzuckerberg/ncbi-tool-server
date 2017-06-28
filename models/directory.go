package models

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"ncbi-tool-server/utils"
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

// GetLatest gets the latest directory listing for the path.
func (d *Directory) GetLatest(pathName string,
	output string) ([]Entry, error) {
	// Setup
	var err error
	resp := []Entry{}
	file := NewFile(d.ctx)
	url := ""

	// Get listing from S3
	listing, err := d.ListObj(pathName)
	if err != nil || len(listing) == 0 {
		return resp, errors.New("empty or non-existent directory")
	}

	// Process results
	for _, val := range listing {
		key := *val.Key
		if output == "with-URLs" {
			url, err = file.keyToURL(key)
			if err != nil {
				return resp, err
			}
		}
		entry := Entry{Path: *val.Key, URL: url}
		resp = append(resp, entry)
	}

	if len(resp) == 0 {
		err = errors.New("no results")
	}
	return resp, err
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
	for _, val := range listing {
		key := file.getS3Key(val)
		if output == "with-URLs" {
			url, err = file.keyToURL(key)
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
	query := fmt.Sprintf("select e.PathName, e.VersionNum, "+
		"e.DateModified, e.ArchiveKey "+
		"from entries as e "+
		"inner join ( "+
		"select max(VersionNum) VersionNum, PathName "+
		"from entries "+
		"where PathName LIKE '%s%%' "+
		"and DateModified <= '%s' "+
		"group by PathName ) as max "+
		"on max.PathName = e.PathName "+
		"and max.VersionNum = e.VersionNum",
		pathName, inputTime)
	rows, err := d.ctx.Db.Query(query)
	if err != nil {
		return res, errors.New("no results found")
	}
	defer rows.Close()

	// Process results
	for rows.Next() {
		md := Metadata{}
		err = rows.Scan(&md.Path, &md.Version, &md.ModTime, &md.ArchiveKey)
		if err != nil {
			return res, err
		}
		res = append(res, md)
	}
	return res, err
}

// ListObj lists objects with a given prefix in S3. Lists the files
// in a S3 folder path.
func (d *Directory) ListObj(pathName string) ([]*s3.Object, error) {
	var err error
	// Remove leading forward slash
	pathName = pathName[1:]
	params := &s3.ListObjectsInput{
		Bucket: aws.String(d.ctx.Bucket),
		Prefix: aws.String(pathName),
	}

	res, err := d.ctx.Store.ListObjects(params)

	// Filter out zero size objects to ignore the 'folder' objects
	pruned := []*s3.Object{}
	for _, val := range res.Contents {
		// Don't include the query itself
		if int(*val.Size) > 0 && *val.Key != pathName {
			pruned = append(pruned, val)
		}
	}
	return pruned, err
}
