package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sethgrid/multibar"
)

var me *user.User

type Command interface {
	Init(*flag.FlagSet, *log.Logger)
	Run(*session.Session) error
}

var commands = map[string]Command{
	"s3create": new(s3createCmd),
	"s3fill":   new(s3fillCmd),
	"s3ls":     new(s3ls),
}

const usage = `amz COMMAND [ARGS...]

Available commands are:

	s3create -help
	s3fill   -help
	s3ls     -help
`

func matches(err error, code string) bool {
	switch e := err.(type) {
	case awserr.Error:
		return strings.Contains(strings.ToLower(e.Code()), code)
	default:
		return false
	}
}

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func init() {
	var err error
	if me, err = user.Current(); err != nil {
		die(err)
	}
	rand.Seed(time.Now().UnixNano() + int64(os.Getpid()))
}

func main() {
	if len(os.Args) == 1 {
		die(usage)
	}
	name, args := os.Args[1], os.Args[2:]
	cmd, ok := commands[name]
	if !ok {
		die(usage)
	}
	f := flag.NewFlagSet("amz", flag.ContinueOnError)
	l := log.New(os.Stderr, "["+name+"] ", log.LstdFlags)
	region := f.String("region", "eu-central-1", "Region name.")
	cmd.Init(f, l)
	if err := f.Parse(args); err != nil {
		die(err)
	}
	cfg := &aws.Config{
		Credentials: credentials.NewCredentials(&credentials.EnvProvider{}),
		Region:      region,
	}
	if err := cmd.Run(session.New(cfg)); err != nil {
		die(err)
	}
}

type s3ls struct {
	N      int
	Path   string
	Bucket string
	Log    *log.Logger
}

func (cmd *s3ls) Init(flags *flag.FlagSet, log *log.Logger) {
	flags.IntVar(&cmd.N, "n", 0, "List max n objects.")
	flags.StringVar(&cmd.Bucket, "bucket", "amz-bucket-"+me.Username, "Bucket name.")
	flags.StringVar(&cmd.Path, "path", "", "Relative path within bucket.")
	cmd.Log = log
}

func (cmd *s3ls) Run(session *session.Session) error {
	svc := s3.New(session)
	params := &s3.ListObjectsInput{
		Bucket: aws.String(cmd.Bucket),
	}
	if cmd.Path != "" {
		params.Prefix = aws.String(cmd.Path + "/")
	}
	return svc.ListObjectsPages(params, func(resp *s3.ListObjectsOutput, _ bool) bool {
		for _, obj := range resp.Contents {
			fmt.Println(aws.StringValue(obj.Key))
		}
		return true
	})
}

type s3createCmd struct {
	Bucket string
	Log    *log.Logger
}

func (cmd *s3createCmd) Init(flags *flag.FlagSet, log *log.Logger) {
	flags.StringVar(&cmd.Bucket, "bucket", "amz-bucket-"+me.Username, "Bucket name.")
	cmd.Log = log
}

func (cmd *s3createCmd) Run(session *session.Session) error {
	svc := s3.New(session)
	params := &s3.CreateBucketInput{
		Bucket: aws.String(cmd.Bucket),
		ACL:    aws.String(s3.BucketCannedACLPrivate),
	}
	_, err := svc.CreateBucket(params)
	if err != nil {
		return err
	}
	return svc.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: aws.String(cmd.Bucket)})
}

type s3fillCmd struct {
	N      int
	Path   string
	Bucket string
	Log    *log.Logger
}

func (cmd *s3fillCmd) Init(flags *flag.FlagSet, log *log.Logger) {
	flags.IntVar(&cmd.N, "n", 1000, "Add n objects under random names.")
	flags.StringVar(&cmd.Bucket, "bucket", "amz-bucket-"+me.Username, "Bucket name.")
	flags.StringVar(&cmd.Path, "path", "", "Relative path within bucket.")
	cmd.Log = log
}

func (cmd *s3fillCmd) Run(session *session.Session) error {
	svc := s3.New(session)
	bars, err := multibar.New()
	if err != nil {
		return err
	}
	go bars.Listen()
	progress := bars.MakeBar(cmd.N, cmd.Bucket)
	left := cmd.N
	for left > 0 {
		s := fmt.Sprintf("object-%d", rand.Int63())
		key := path.Join(cmd.Path, s)
		params := &s3.PutObjectInput{
			Bucket:        aws.String(cmd.Bucket),
			ACL:           aws.String(s3.BucketCannedACLPrivate),
			Key:           aws.String(key),
			Body:          strings.NewReader(s),
			ContentLength: aws.Int64(int64(len(s))),
			ContentType:   aws.String("text/plain"),
		}
		_, err := svc.PutObject(params)
		if matches(err, "duplicate") {
			cmd.Log.Printf("bucket=%q, key=%q: %s", cmd.Bucket, key, err)
			continue
		}
		if err != nil {
			return err
		}
		left--
		progress(cmd.N - left - 1)
	}
	return nil
}
