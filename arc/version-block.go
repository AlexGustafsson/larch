package arc

import (
  "time"
  "bufio"
  "fmt"
)

const DateTimeFormat = "20060102030405"

// https://archive.org/web/researcher/ArcFileFormat.php

// "version"
// filedesc://<path><sp><version specific data><sp><length><nl>
// <version-number><sp><reserved><sp><origin-code><nl>
// <URL-record-definition><nl>
// <nl>

// "version 1"
// filedesc://<path><sp><ip_address><sp><date><sp>text/plain<sp><length><nl>
// 1<sp><reserved><sp><origin-code><nl>
// <URL IP-address ArchivArchivee-date Content-type Archive-length<nl>
// <nl>

// VersionBlockV1 is a version block according to the ARC v1 specification
type VersionBlockV1 struct {
  // Path is the original path name of the archive file
  Path string
  // IP is the address is the address of the machine that created the archive file
  IP string
  // Date is the date the archive file was created
  Date time.Time
  // Length specifies the size, in bytes, of the rest of the version block
  Length uint64
  // Reserved is any single-word avlue
  Reserved string
  // OriginCode is the name of the gathering organization with no white space
  OriginCode string
  // URLRecord is names of fields in URL records
  URLRecord string
}

// <url><sp>
// <ip-address><sp>
// <archive-date><sp>
// <content-type><sp>
// <length><nl>
type URLRecordV1 struct {

}

// VersionBlockV2 is a version block according to the ARC v2 specification
type VersionBlockV2 struct {
  // Path is the original path name of the archive file
  Path string
  // IP is the address is the address of the machine that created the archive file
  IP string
  // Date is the date the archive file was created
  Date time.Time
  // FileName is the name of the file
  FileName string
  // Length specifies the size, in bytes, of the rest of the version block
  Length uint64
  // Reserved is any single-word avlue
  Reserved string
  // OriginCode is the name of the gathering organization with no white space
  OriginCode string
  // URLRecord is names of fields in URL records
  URLRecord string
}

// <url><sp>
// <ip-address><sp>
// <archive-date><sp>
// <content-type><sp>
// <result-code><sp>
// <checksum><sp>
// <location><sp>
// <offset><sp>
// <filename><sp>
// <length><nl>
type URLRecordV2 struct {

}

// "version 1"
// filedesc://<path><sp><ip_address><sp><date><sp>text/plain<sp><length><nl>
// 1<sp><reserved><sp><origin-code><nl>
// <URL IP-address ArchivArchivee-date Content-type Archive-length<nl>
// <nl>
func (block *VersionBlockV1) String() string {
  return fmt.Sprintf(
    "filedesc://%s %s %s text/plain %d\r\n1 %s %s\r\n%s\n",
    block.Path,
    block.IP,
    block.Date.Format(DateTimeFormat),
    block.Length,
    block.Reserved,
    block.OriginCode,
    block.URLRecord.String(),
  )
}

// "version 2"
// filedesc://<path><sp><ip_address><sp><date><sp>text/plain<sp>200<sp>-<sp>-<sp>0<sp><filename><sp><length><nl>
// 2<sp><reserved><sp><origin-code><nl>
// URL<sp>IP-address<sp>Archive-date<sp>Content-type<sp>Result-code<sp>Checksum<sp>Location<sp> Offset<sp>Filename<sp>Archive-length<nl>
// <nl>
func (block *VersionBlockV2) String() string {
  return fmt.Sprintf(
    "filedesc://%s %s %s text/plain 200 - - 0 %s %d\r\n2 %s %s\r\n%s\r\n\r\n",
    block.Path,
    block.IP,
    block.Date.Format(DateTimeFormat),
    block.FileName,
    block.Length,
    block.Reserved,
    block.OriginCode,
    block.URLRecord.String(),
  )
}

func consumeProtocol(reader: io.Reader) err {
  var buffer[11]byte
  _, err := io.ReadFull(reader, buffer)
  if err != nil {
    return err
  }

  if string(buffer) != "filedesc://" {
    return fmt.Errof("Expected filedesc://")
  }

  return nil
}

func ParseVersionBlock(reader: bufio.Reader) (*VersionBlock, err) {
  scanner := bufio.NewScanner(reader)
  scanner.Split(bufio.ScanWords)

  versionBlock := &VersionBlock {}

  // Parse protocol: filedesc://<path>
  if ok := scanner.Scan(); ok {
    if scanner.Text().StartsWith("filedesc://") {
      versionBlock.Path = scanner.Text()[11:]
    } else {
      return nil, fmt.Errorf("Expected filedesc://")
    }
  } else {
    return nil, fmt.Errorf("Unable to get protocol and path")
  }

  // Parse IP: <ip>
  if ok := scanner.Scan(); ok {
    versionBlock.IP = net.ParseIP(scanner.Text())
    if versionBlock.IP == nil {
      return nil, fmt.Errorf("Unable to parse IP address")
    }
  } else {
    return nil, fmt.Errorf("Unable to get IP")
  }

  // Parse date: yyyyMMddhhmmss
  if ok := scanner.Scan(); ok {
  	time, error := time.Parse(DateTimeFormat, scanner.Text())
    if err == nil {
      versionBlock.Date = time
    } else {
      return nil, fmt.Errorf("Unable to parse date")
    }
  } else {
    return nil, fmt.Errorf("Unable to get IP")
  }

  // Parse MIME: always text/plain
  if ok := scanner.Scan(); ok {
    if scanner.Text() != "text/plain" {
      return nil, fmt.Errorf("Got unexpected MIME type")
    }
  } else {
    return nil, fmt.Errorf("Unable to get MIME type")
  }

  // Parse MIME: always text/plain
  if ok := scanner.Scan(); ok {
    if scanner.Text() != "text/plain" {
      return nil, fmt.Errorf("Got unexpected MIME type")
    }
  } else {
    return nil, fmt.Errorf("Unable to get MIME type")
  }

  return versionBlock, nil
}

The Version Block
The version block identifies the original filename, file version, and URL record fields of the archive file.
