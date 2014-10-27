s3put is a simple no-dependence utility to upload files to Amazon S3 buckets.

	Usage: s3put [flags] <filenames to upload>
	  -acl="private": ACL (S3_ACL variable)
	  -ak="": access key (S3_ACCESS_KEY variable)
	  -b="": bucket (S3_BUCKET variable)
	  -p="": prefix path to add to uploaded filename (subdirectory)
	  -reg="us-west-1": region (S3_REGION variable)
	  -sk="": secret key (S3_SECRET_KEY variable)
