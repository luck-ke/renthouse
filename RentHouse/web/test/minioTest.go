package main

//
//func main() {
//	ctx := context.Background()
//	endpoint := "127.0.0.1:9000"
//	accessKeyID := "minioadmin"
//	secretAccessKey := "minioadmin"
//	useSSL := false
//
//	// Initialize minio client object.
//	minioClient, err := minio.New(endpoint, &minio.Options{
//		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
//		Secure: useSSL,
//	})
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	// Make a new bucket called mymusic.
//	bucketName := "image"
//	location := "us-east-1"
//
//	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
//	if err != nil {
//		// Check to see if we already own this bucket (which happens if you run this twice)
//		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
//		if errBucketExists == nil && exists {
//			log.Printf("We already own %s\n", bucketName)
//		} else {
//			log.Fatalln(err)
//		}
//	} else {
//		log.Printf("Successfully created %s\n", bucketName)
//	}
//
//	// Upload the zip file
//	objectName := "flower.png"
//	filePath := "D:\\miniostorage\\flower.png"
//	contentType := "image/png"
//
//	// Upload the zip file with FPutObject
//	info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
//	if err != nil {
//		fmt.Println("失败")
//		log.Fatalln(err)
//	}
//	//fmt.Println(info)
//	//minioClient.PresignedGetObject()
//	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
//}
