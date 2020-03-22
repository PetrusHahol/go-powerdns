package powerdns

import (
	"fmt"

	"github.com/joeig/go-powerdns/v2/lib"
)

// RecordsService handles communication with the records related methods of the Client API
type RecordsService service

// Add creates a new resource record
func (r *RecordsService) Add(domain string, name string, recordType lib.RRType, ttl uint32, content []string) error {
	return r.Change(domain, name, recordType, ttl, content)
}

// Change replaces an existing resource record
func (r *RecordsService) Change(domain string, name string, recordType lib.RRType, ttl uint32, content []string) error {
	rrset := new(lib.RRset)
	rrset.Name = &name
	rrset.Type = &recordType
	rrset.TTL = &ttl
	rrset.ChangeType = lib.ChangeTypePtr(lib.ChangeTypeReplace)
	rrset.Records = lib.RecordSlicePtr(make([]lib.Record, 0))

	for _, c := range content {
		r := lib.Record{Content: lib.StringPtr(c), Disabled: lib.BoolPtr(false), SetPTR: lib.BoolPtr(false)}
		*rrset.Records = append(*rrset.Records, r)
	}

	return r.patchRRset(domain, *rrset)
}

// Delete removes an existing resource record
func (r *RecordsService) Delete(domain string, name string, recordType lib.RRType) error {
	rrset := new(lib.RRset)
	rrset.Name = &name
	rrset.Type = &recordType
	rrset.ChangeType = lib.ChangeTypePtr(lib.ChangeTypeDelete)

	return r.patchRRset(domain, *rrset)
}

func (r *RecordsService) patchRRset(domain string, rrset lib.RRset) error {
	rrset.Name = lib.StringPtr(lib.MakeDomainCanonical(*rrset.Name))

	lib.FixRRset(&rrset)

	payload := lib.RRsets{
		Sets: lib.RRsetSlicePtr([]lib.RRset{
			rrset,
		}),
	}

	req, err := r.client.newRequest("PATCH", fmt.Sprintf("servers/%s/zones/%s", r.client.VHost, lib.TrimDomain(domain)), nil, payload)
	if err != nil {
		return err
	}

	_, err = r.client.do(req, nil)

	return err
}
