package models

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"ncbi-tool-server/utils"
	"path"
	"strconv"
	"time"
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
func (f *File) versionFromTime(path string, inputTime string) (Metadata,
	error) {
	res := f.ctx.Db.QueryRow("select * from entries where "+
		"PathName=? and DateModified <= ? order "+
		"by VersionNum desc limit 1", path, inputTime)
	return rowToMetadata(res)
}

// Gets the metadata of the specified or latest version of the file.
func (f *File) entryFromVersion(path string, version int) (Metadata, error) {
	var res *sql.Row
	if version > 0 {
		// Get specified version
		res = f.ctx.Db.QueryRow("select * from entries "+
			"where PathName=? and VersionNum=?", path, version)
	} else {
		// Get latest version
		res = f.ctx.Db.QueryRow("select * from entries "+
			"where PathName=? order by VersionNum desc limit 1", path)
	}
	return rowToMetadata(res)
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
		ResponseContentDisposition: aws.String("attachment; filename=" +
			downloadName),
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
	rows, err := f.ctx.Db.Query("select * from entries "+
		"where PathName=? order by VersionNum desc", path)
	if err != nil {
		return res, err
	}
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			err = utils.ComboErr("Couldn't close db rows.", closeErr, err)
			log.Println(err)
		}
	}()

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

// RowToMetadata converts a SQL row into a Metadata entry and handles errors.
func rowToMetadata(row *sql.Row) (Metadata, error) {
	md := Metadata{}
	err := row.Scan(&md.Path, &md.Version, &md.ModTime, &md.ArchiveKey)
	switch {
	case err == sql.ErrNoRows:
		err = utils.NewErr("No results for this query.", err)
		log.Print(err)
	case err != nil:
		err = utils.NewErr("Error retrieving results.", err)
		log.Print(err)
	}
	return md, err
}
