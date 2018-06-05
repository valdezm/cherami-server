// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cassandra

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

type (
	// CQLClient is the interface for implementations
	// that provide way to talk to cassandra through CQL
	CQLClient interface {
		// Exec executes the cql statement
		Exec(stmt string) error
		// ReadSchemaVersion returns the current schema version for the keyspace
		ReadSchemaVersion() (int64, error)
		// UpdateSchemaVersion updates the schema version for the keyspace
		UpdateSchemaVersion(newVersion int64, minCompatibleVersion int64) error
		// WriteSchemaUpdateLog adds an entry to the schema update history table
		WriteSchemaUpdateLog(oldVersion int64, newVersion int64, manifestMD5 string, desc string) error
	}
	cqlClient struct {
		session       *gocql.Session
		clusterConfig *gocql.ClusterConfig
	}
)

var errNoHosts = errors.New("Cassandra hosts list is empty or malformed")
var errGetSchemaVersion = errors.New("Failed to get current schema version from cassandra")

const (
	newLineDelim       = '\n'
	defaultTimeout     = 30 * time.Second
	defaultConsistency = "QUORUM" // schema updates must always be QUORUM
)

const (
	readSchemaVersionCQL        = `SELECT curr_version from schema_version where keyspace_name=?`
	writeSchemaVersionCQL       = `INSERT into schema_version(keyspace_name, creation_time, curr_version, min_compatible_version) VALUES (?,?,?,?)`
	writeSchemaUpdateHistoryCQL = `INSERT into schema_update_history(year, month, update_time, old_version, new_version, manifest_md5, description) VALUES(?,?,?,?,?,?,?)`
)

// newCQLClient returns a new instance of CQLClient
func newCQLClient(config *SchemaUpdaterConfig) (CQLClient, error) {
	hosts := parseHosts(config.HostsCsv)
	if len(hosts) == 0 {
		return nil, errNoHosts
	}
	clusterCfg := gocql.NewCluster(hosts...)
	clusterCfg.Port = config.Port
	clusterCfg.Keyspace = config.Keyspace
	clusterCfg.Timeout = defaultTimeout
	clusterCfg.ProtoVersion = config.ProtoVersion
	clusterCfg.Consistency = gocql.ParseConsistency(defaultConsistency)

	if config.Username != "" && config.Password != "" {
		clusterCfg.Authenticator = gocql.PasswordAuthenticator{
			Username: config.Username,
			Password: config.Password,
		}
	}

	cqlClient := new(cqlClient)
	cqlClient.clusterConfig = clusterCfg
	var err error
	cqlClient.session, err = clusterCfg.CreateSession()
	if err != nil {
		return nil, err
	}
	return cqlClient, nil
}

func (client *cqlClient) ReadSchemaVersion() (int64, error) {
	query := client.session.Query(readSchemaVersionCQL, client.clusterConfig.Keyspace)
	iter := query.Iter()
	var version int64
	if !iter.Scan(&version) {
		iter.Close()
		return 0, errGetSchemaVersion
	}
	if err := iter.Close(); err != nil {
		return 0, err
	}
	return version, nil
}

func (client *cqlClient) UpdateSchemaVersion(newVersion int64, minCompatibleVersion int64) error {
	query := client.session.Query(writeSchemaVersionCQL, client.clusterConfig.Keyspace, time.Now(), newVersion, minCompatibleVersion)
	return query.Exec()
}

func (client *cqlClient) WriteSchemaUpdateLog(oldVersion int64, newVersion int64, manifestMD5 string, desc string) error {
	now := time.Now().UTC()
	query := client.session.Query(writeSchemaUpdateHistoryCQL)
	query.Bind(now.Year(), int(now.Month()), now, oldVersion, newVersion, manifestMD5, desc)
	return query.Exec()
}

func (client *cqlClient) Exec(stmt string) error {
	return client.session.Query(stmt).Exec()
}

func parseHosts(input string) []string {
	var hosts = make([]string, 0)
	for _, h := range strings.Split(input, ",") {
		if host := strings.TrimSpace(h); len(host) > 0 {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

// ParseCQLFile takes a cql file path as input
// and returns an array of cql statements on
// success.
func ParseCQLFile(filePath string) ([]string, error) {

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(f)

	var line string
	var currStmt string
	var stmts = make([]string, 0, 4)

	for err == nil {

		line, err = reader.ReadString(newLineDelim)
		line = strings.TrimSpace(line)
		if len(line) < 1 {
			continue
		}

		// Filter out the comment lines, the
		// only recognized comment line format
		// is any line that starts with double dashes
		tokens := strings.Split(line, "--")
		if len(tokens) > 0 && len(tokens[0]) > 0 {
			currStmt += tokens[0]
			// semi-colon is the end of statement delim
			if strings.HasSuffix(currStmt, ";") {
				stmts = append(stmts, currStmt)
				currStmt = ""
			}
		}
	}

	if err == io.EOF {
		return stmts, nil
	}

	return nil, err
}
