package warc

import (
	"bytes"
	"io"
	"time"
)

const (
	// TypeInfo is a type of record that contains information about the file.
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
	TruncatedReasonTime = "time"
	// TruncatedReasonDisconnect is the reason for a record to be truncated due to a network disconnect.
	TruncatedReasonDisconnect = "disconnect"
	// TruncatedReasonUnspecified is the reason for a record to be truncated due to other or unknown issues.
	TruncatedReasonUnspecified = "unspecified"
)

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
	ContentType string `warc:"ContentType,omitempty"`
	// ConcurrentTo (WARC-Concurrent-To) is the Record IDs of any records created as part of the same capture as the current record. Must be a valid URI.
	ConcurrentTo string `warc:"WARC-Concurrent-To,omitempty"`
	// BlockDigest (WARC-Block-Digest) is a digest of the full block of the record. Format algorithm:digest.
	BlockDigest string `warc:"WARC-Block-Digest,omitempty"`
	// PayloadDigest  (WARC-Payload-Digest) is a digest of the payload of the record. Format algorithm:digest.
	PayloadDigest string `warc:"WARC-Payload-Digest,omitempty"`
	// IPAddress (WARC-IP-Address) is the IPv4 or IPv6 address of the server giving the response.
	IPAddress string `warc:"WARC-IP-Address,omitempty"`
	// RefersTo (WARC-Refers-To) is the Record ID of a single record for which the present record holds additional content.
	RefersTo string `warc:"WARC-Refers-To,omitempty"`
	// TargetURI (WARC-Target-URI) is the original URI whose capture gave rise to the information content in this record.
	TargetURI string `warc:"WARC-Target-URI,omitempty"`
	// Truncated (WARC-Truncated) is the reason a record payload was truncated.
	Truncated string `warc:"WARC-Truncated,omitempty"`
	// InfoID (WARC-Warcinfo-ID) indicates the Record ID of the associated "warcinfo" record for this record.
	InfoID string `warc:"WARC-Warcinfo-ID,omitempty"`
	// Filename (WARC-Filename) is the filename containing the current "warcinfo" record.
	Filename string `warc:"WARC-Filename,omitempty"`
	// Profile (WARC-Profile) is a URI signifying the kind of analysis and handling applied in a "revisit" record.
	Profile string `warc:"WARC-Profile,omitempty"`
	// IdentifiedPayloadType (WARC-Identified-Payload-Type) is the content-type of the record's payload as determined by an independent check.
	IdentifiedPayloadType string `warc:"WARC-Identified-Payload-Type,omitempty"`
	// SegmentNumber (WARC-Segment-Number) reports the current record's relative ordering in a sequence of segmented records. Mandatory for "continuation" records.
	SegmentNumber uint64 `warc:"WARC-Segment-Number"`
	// SegmentOriginID (WARC-Segment-Origin-ID) identifies the starting record in a series of segmented records.
	SegmentOriginID string `warc:"WARC-Segment-Origin-ID,omitempty"`
	// SegmentTotalLength (WARC-Segment-Total-Length) reports the total length of all segment content blocks when concatenated together.
	SegmentTotalLength uint64 `warc:"WARC-Segment-Total-Length"`
}

// ReadHeader reads the header of a record.
func ReadHeader(reader io.ReadSeeker) (*Header, error) {
	header := &Header{}
	return header, nil
}

// Write writes the header to a stream.
func (header *Header) Write(writer io.Writer) error {
	data, err := Marshal(header)
	if err != nil {
		return err
	}

	writer.Write(data)

	return nil
}

// String converts the header into a string
func (header *Header) String() (string, error) {
	buffer := new(bytes.Buffer)
	err := header.Write(buffer)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
