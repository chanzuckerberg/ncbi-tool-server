package models

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"ncbi-tool-server/utils"
	"strconv"
	"time"
	"path"
)

// File Model
type File struct {
	ctx *utils.Context
}

// NewFile returns a new file instance
func NewFile(ctx *utils.Context) *File {
	return &File{
		ctx: ctx,
	}
}

// Metadata about a file version from the db
type Metadata struct {
	Path       string
	Version    int
	ModTime    sql.NullString
	ArchiveKey sql.NullString
}

// Entry contains info about a file version entry for formatting
type Entry struct {
	Path    string `json:",omitempty"`
	Version int    `json:",omitempty"`
	ModTime string `json:",omitempty"`
	URL     string `json:",omitempty"`
}

// GetVersion gets the response for a file and version.
func (f *File) GetVersion(path string,
	version string) (Entry, error) {
	num, _ := strconv.Atoi(version)
	info, err := f.entryFromVersion(path, num)
	if err != nil {
		return Entry{}, err
	}
	return f.entryFromMetadata(info)
}

// GetLatest gets the response for a file, latest version.
func (f *File) GetLatest(path string) (Entry, error) {
	info, err := f.entryFromVersion(path, 0)
	if err != nil {
		return Entry{}, err
	}
	return f.entryFromMetadata(info)
}

// GetAtTime gets the file version at/just before the given time.
func (f *File) GetAtTime(path string,
	inputTime string) (Entry, error) {
	info, err := f.versionFromTime(path, inputTime)
	if err != nil {
		return Entry{}, err
	}
	return f.entryFromMetadata(info)
}

// Gets an Entry for a file from the metadata information.
func (f *File) entryFromMetadata(info Metadata) (Entry, error) {
	key := f.getS3Key(info)
	downloadName := path.Base(info.Path)
	url, err := f.keyToURL(key, downloadName)
	if err != nil {
		return Entry{}, err
	}
	return Entry{
		info.Path,
		info.Version,
		info.ModTime.String,
		url}, err
}

// Gets metadata entry based on file name and given time.
// Finds the version of the file just before the given time, if any.
func (f *File) versionFromTime(path string,
	inputTime string) (Metadata, error) {
	// Query the database
	query := fmt.Sprintf("select * from entries where "+
		"PathName='%s' and DateModified <= '%s' order "+
		"by VersionNum desc", path, inputTime)
	log.Print("Query: " + query)
	return f.topFromQuery(query)
}

// Gets column info from the top db result of the query.
// Runs the query and return the columns of just the first result.
func (f *File) topFromQuery(query string) (Metadata, error) {
	md := Metadata{}
	row, err := f.ctx.Db.Query(query)
	if row == nil || err != nil {
		return md, err
	}
	defer row.Close()

	// Process results
	if !row.Next() {
		return md, errors.New("no results for this query")
	}
	err = row.Scan(&md.Path, &md.Version, &md.ModTime, &md.ArchiveKey)
	return md, err
}

// Gets the metadata of the specified or latest version of the file.
func (f *File) entryFromVersion(path string,
	version int) (Metadata, error) {
	query := ""
	if version > 0 {
		// Get specified version
		query = fmt.Sprintf("select * from entries "+
			"where PathName='%s' and VersionNum=%d", path, version)
	} else {
		// Get latest version
		query = fmt.Sprintf("select * from entries "+
			"where PathName='%s' order by VersionNum desc", path)
	}
	return f.topFromQuery(query)
}

// Gets the S3 key for the given entry.
func (f *File) getS3Key(info Metadata) string {
	if !info.ArchiveKey.Valid {
		// VersionEntry is there but not archived. Just serve the latest.
		return info.Path
	}
	// Make the archive folder path
	archiveKey := info.ArchiveKey.String
	return fmt.Sprintf("/archive/%s", archiveKey)
}

// Gets a pre-signed temporary URL from S3 for a key.
// Serves back link for client downloads.
func (f *File) keyToURL(key string, downloadName string) (string, error) {
	req, _ := f.ctx.Store.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(f.ctx.Bucket),
		Key:    aws.String(key),
		ResponseContentDisposition: aws.String("attachment; filename="+downloadName),
	})

	out, err := req.Presign(1 * time.Hour)
	if err != nil {
		log.Println(out)
		return "", errors.New("Couldn't generate URL. " + err.Error())
	}
	return out, err
}

// GetHistory gets the revision history of a file. Gets list of
// versions and modTimes.
func (f *File) GetHistory(path string) ([]Entry, error) {
	var err error
	res := []Entry{}

	// Query the database
	query := fmt.Sprintf("select * from entries "+
		"where PathName='%s' order by VersionNum desc", path)
	rows, err := f.ctx.Db.Query(query)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	// Process results
	md := Metadata{}
	for rows.Next() {
		err = rows.Scan(&md.Path, &md.Version,
			&md.ModTime, &md.ArchiveKey)
		if err != nil {
			return res, err
		}
		entry := Entry{
			Path:    md.Path,
			Version: md.Version,
			ModTime: md.ModTime.String,
		}
		res = append(res, entry)
	}
	return res, err
}
