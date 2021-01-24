package warc

import (
	"fmt"
	"io"
	"time"
)

// See http://bibnum.bnf.fr/WARC/WARC_ISO_28500_version1_latestdraft.pdf

const (
	TypeInfo string = "warcinfo"
	// TypeResponse is a type of record that is a complete schema-specific response, including network protocol information where possible.
	TypeResponse = "response"
	// TypeResource is a type of record that holds a aesource, without full protocol response information, such as a file locally accessible.
	TypeResource = "resource"
	// TypeRequest is a type of record that holds the details of a complete schema-specific request, including network protocol information where possible.
	TypeRequest = "request"
	// TypeMetadata is a type of record that ...
	TypeMetadata = "metadata"
	// TypeRevisit is a type of record that describes the revisitation of content already archived.
	TypeRevisit = "revisit"
	// TypeConversion is a type of record that describes an alternative version of another record's content that was created during the archival process.
	TypeConversion = "conversion"
	// TypeContinuation is a type of a record appended to corresponding prior record block(s) to create a logically complete full-sized original record.
	TypeContinuation = "continuation"
)

const (
	// TruncatedReasonLength is the reason for a record to be truncated due to a record exceeding a configured max length.
	TruncatedReasonLength string = "length"
	// TruncatedReasonTime is the reason for a record to be truncated due to a process exceeding a configured max time.
	TrucatedReasonTime = "time"
	// TruncatedReasonDisconnect is the reason for a record to be truncated due to a network disconnect.
	TrucatedReasonDisconnect = "disconnect"
	// TruncatedReasonUnspecified is the reason for a record to be truncated due to other or unknown issues.
	TrucatedReasonUnspecified = "unspecified"
)

// File is a WARC file containing one or more records.
type File struct {
	// Records are all the records contained within the file.
	Records []*Record
}

// Record is a WARC record.
type Record struct {
	// Header is a WARC header containing information about the payload. Guaranteed to exist.
	Header *Header
	// Payload is the body of the record. May not exist.
	Payload *Payload
}

// Header is a WARC header containing information about the payload.
type Header struct {
	// Type (WARC-Type). Mandatory.
	Type string `warc:"WARC-Type"`
	// RecordID (WARC-Record-ID) is a globally unique identifier. Must be a valid URI. Mandatory.
	RecordID string `warc:"WARC-Record-ID"`
	// Date (WARC-Date) is the time at which the data capture for the record began. Mandatory.
	Date time.Time `warc:"WARC-Date"`
	// ContentLength is the number of octents in the block. If no block is present 0 is used. Mandatory.
	ContentLength uint64 `warc:"ContentLength"`
	// ContentType is the RFC2045 MIME type of the information in the record's block. Mandatory for non-empty, non-continuation records.
	ContentType string `warc:"ContentLength"`
	// ConcurrentTo (WARC-Concurrent-To) is the Record IDs of any records created as part of the same capture as the current record. Must be a valid URI.
	ConcurrentTo string `warc:"WARC-Concurrent-To"`
	// BlockDigest (WARC-Block-Digest) is a digest of the full block of the record. Format algorithm:digest.
	BlockDigest string `warc:"WARC-Block-Digest"`
	// PayloadDigest  (WARC-Payload-Digest) is a digest of the payload of the record. Format algorithm:digest.
	PayloadDigest string `warc:"WARC-Payload-Digest"`
	// IPAddress (WARC-IP-Address) is the IPv4 or IPv6 address of the server giving the response.
	IPAddress string `warc:"WARC-IP-Address"`
	// RefersTo (WARC-Refers-To) is the Record ID of a single record for which the present record holds additional content.
	RefersTo string `warc:"WARC-Refers-To"`
	// TargetURI (WARC-Target-URI) is the original URI whose capture gave rise to the information content in this record.
	TargetURI string `warc:"WARC-Target-URI"`
	// Truncated (WARC-Truncated) is the reason a record payload was truncated.
	Truncated string `warc:"WARC-Truncated"`
	// InfoID (WARC-Warcinfo-ID) indicates the Record ID of the associated "warcinfo" record for this record.
	InfoID string `warc:"WARC-Warcinfo-ID"`
	// Filename (WARC-Filename) is the filename containing the current "warcinfo" record.
	Filename string `warc:"WARC-Filename"`
	// Profile (WARC-Profile) is a URI signifying the kind of analysis and handling applied in a "revisit" record.
	Profile string `warc:"WARC-Profile"`
	// IdentifiedPayloadType (WARC-Identified-Payload-Type) is the content-type of the record's payload as determined by an independent check.
	IdentifiedPayloadType string `warc:"WARC-Identified-Payload-Type"`
	// SegmentNumber (WARC-Segment-Number) reports the current record's relative ordering in a sequence of segmented records. Mandatory for "continuation" records.
	SegmentNumber uint64 `warc:"WARC-Segment-Number"`
	// SegmentOriginID (WARC-Segment-Origin-ID) identifies the starting record in a series of segmented records.
	SegmentOriginID string `warc:"WARC-Segment-Origin-ID"`
	// SegmentTotalLength (WARC-Segment-Total-Length) reports the total length of all segment content blocks when concatenated together.
	SegmentTotalLength uint64 `warc:"WARC-Segment-Total-Length"`
}

// Payload is any payload of a WARC Record
type Payload struct {
	// Data is the raw data of the payload
	Data []byte
	// Length is the length of the data in bytes
	Length uint64
}

// InfoPayload is the payload of a "warcinfo" record.
type InfoPayload struct {
	Payload
	// Operator contains contact information for the operator who created this WARC resource.
	Operator string `warc:"operator"`
	// Software is the software and version used to create this WARC resource.
	Software string `warc:"software"`
	// Robots is the robots policy followed by the harvester creating this WARC resource.
	Robots string `warc:"robots"`
	// Hostname is the hostname of the machine that created this WARC resource.
	Hostname string `warc:"hostname"`
	// The IP address of the machine that created this WARC resource.
	IP string `warc:"ip"`
	// UserAgent is the HTTP 'user-agent' header usually sent by the harvester along with each request.
	UserAgent string `warc:"http-header-user-agent"`
	// From is the HTTP 'From' header usually sent by the harvester along with each request.
	From string `warc:"http-header-from"`
}

// MetadataPayload is a payload record that contains content created in order to further describe a harvested resource.
type MetadataPayload struct {
	Payload
	// Via is the referring URI from which the archived URI was discorvered.
	Via string `warc:"via"`
	// HopsFromSeed describes the type of each hop from a starting URI to the current URI.
	HopsFromSeed string `warc:"hopsFromSeed"`
	// FetchTime is the time that it took to collect the archived URI, starting from the initation of network traffic.
	FetchTime time.Duration `warc:"fetchTimeMs"`
}

// Read reads an entire WARC file.
func Read(reader io.ReadSeeker) (*File, error) {
	// scanner := bufio.NewScanner(reader)

	file := &File{
		Records: make([]*Record, 0),
	}

	for {
		record, err := ReadRecord(reader)
		if err != nil {
			return nil, err
		}

		// No record received
		if record == nil {
			break
		}

		file.Records = append(file.Records, record)
	}

	return file, nil
}

// ReadRecord reads a record. Returns nil for the record if none was read (EOF).
func ReadRecord(reader io.ReadSeeker) (*Record, error) {
	header, err := ReadHeader(reader)
	if err != nil {
		return nil, err
	}

	payload, err := ReadPayload(reader, header)
	if err != nil {
		return nil, err
	}

	record := &Record{
		Header:  header,
		Payload: payload,
	}

	return record, nil
}

// ReadHeader reads the header of a record.
func ReadHeader(reader io.ReadSeeker) (*Header, error) {
	// TODO: Actually read header
	return nil, nil
}

// ReadPayload reads the payload of a record.
func ReadPayload(reader io.ReadSeeker, header *Header) (*Payload, error) {
	if header.ContentLength <= 0 {
		return nil, nil
	}

	payload := &Payload{
		Data:   make([]byte, header.ContentLength),
		Length: header.ContentLength,
	}

	bytesRead, err := io.ReadFull(reader, payload.Data)
	if err != nil {
		return nil, err
	}

	if bytesRead != int(header.ContentLength) {
		return nil, fmt.Errorf("Unable to read payload. Expected %v bytes got %v", header.ContentLength, bytesRead)
	}

	payload, err = parsePayload(payload, header)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// parsePayload parses a single payload if it's of a supported type. Leaves it unchanged otherwise.
func parsePayload(payload *Payload, header *Header) (*Payload, error) {
	if header.Type == TypeInfo {
		return parseInfoPayload(payload, header)
	} else if header.Type == TypeMetadata {
		return parseMetadataPayload(payload, header)
	}

	return payload, nil
}

// parseInfoPayload parses a WARC info record's payload.
func parseInfoPayload(payload *Payload, header *Header) (*Payload, error) {
	// TODO: Actually parse payload
	return payload, nil
}

// parseInfoPayload parses a WARC metadata record's payload.
func parseMetadataPayload(payload *Payload, header *Header) (*Payload, error) {
	// TODO: Actually parse payload
	return payload, nil
}
