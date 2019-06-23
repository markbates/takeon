package here

import "encoding/json"

// Package attempts to gather info for the requested package.
//
// From the `go help list` docs:
//	The -find flag causes list to identify the named packages but not
//	resolve their dependencies: the Imports and Deps lists will be empty.
//
// A workaround for this issue is to use the `Dir` field in the
// returned `Info` value and pass it to the `Dir(string) (Info, error)`
// function to return the complete data.
func Package(p string) (Info, error) {
	i := newInfo()
	b, err := run("go", "list", "-json", "-find", p)
	if err != nil {
		return i, err
	}
	if err := json.Unmarshal(b, &i); err != nil {
		return i, err
	}

	if err := setEnv(&i); err != nil {
		return i, err
	}
	return i, nil
}
