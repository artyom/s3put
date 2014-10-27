package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/artyom/autoflags"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

func main() {
	log.SetFlags(0)
	defaultACL := os.Getenv("S3_ACL")
	if len(defaultACL) == 0 {
		defaultACL = "private"
	}
	reg := os.Getenv("S3_REGION")
	if len(reg) == 0 {
		reg = "us-west-1"
	}
	config := struct {
		ACL       string `flag:"acl,ACL (S3_ACL variable)"`
		Region    string `flag:"reg,region (S3_REGION variable)"`
		Bucket    string `flag:"b,bucket (S3_BUCKET variable)"`
		AccessKey string `flag:"ak,access key (S3_ACCESS_KEY variable)"`
		SecretKey string `flag:"sk,secret key (S3_SECRET_KEY variable)"`
		Prefix    string `flag:"p,prefix path to add to uploaded filename (subdirectory)"`
	}{
		ACL:       defaultACL,
		Region:    reg,
		Bucket:    os.Getenv("S3_BUCKET"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	}
	if err := autoflags.Define(&config); err != nil {
		log.Fatal(err)
	}
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	if len(config.Bucket) == 0 {
		log.Fatal("No bucket name given")
	}
	if len(config.AccessKey) == 0 || len(config.SecretKey) == 0 {
		log.Fatal("Both AccessKey and SecretKey should be set")
	}
	auth := aws.Auth{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	}
	region, ok := aws.Regions[config.Region]
	if !ok {
		log.Printf("Invalid region provided: %q", config.Region)
		log.Print("Supported regions are:")
		for r := range aws.Regions {
			log.Printf("- %s", r)
		}
		os.Exit(1)
	}
	acl, ok := supportedACLs[config.ACL]
	if !ok {
		log.Printf("Invalid ACL provided: %q", config.ACL)
		log.Printf("Supported ACLs are:")
		for r := range supportedACLs {
			log.Printf("- %s", r)
		}
		os.Exit(1)
	}

	log.SetFlags(log.Ltime)
	connection := s3.New(auth, region)
	bucket := connection.Bucket(config.Bucket)
	buf := make([]byte, 512)
	for _, fn := range flag.Args() {
		f, err := os.Open(fn)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		st, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		n, err := f.Read(buf)
		switch err {
		case nil, io.EOF:
		default:
			log.Fatal(err)
		}
		contType := http.DetectContentType(buf[:n])
		if _, err := f.Seek(0, os.SEEK_SET); err != nil {
			log.Fatal(err)
		}
		log.Printf("uploading %s (%s)", f.Name(), contType)
		err = bucket.PutReader(
			path.Join(config.Prefix, path.Base(f.Name())),
			f, st.Size(), contType, acl,
		)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <filenames to upload>\n", os.Args[0])
		flag.PrintDefaults()
	}
}

var supportedACLs = map[string]s3.ACL{
	"private":                   s3.Private,
	"public-read":               s3.PublicRead,
	"public-read-write":         s3.PublicReadWrite,
	"authenticated-read":        s3.AuthenticatedRead,
	"bucket-owner-read":         s3.BucketOwnerRead,
	"bucket-owner-full-control": s3.BucketOwnerFull,
}
