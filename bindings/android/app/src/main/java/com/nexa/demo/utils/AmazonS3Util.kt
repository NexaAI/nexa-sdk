package com.nexa.demo.utils

import android.content.Context
import android.util.Log
import com.amazonaws.auth.AnonymousAWSCredentials
import com.amazonaws.auth.BasicAWSCredentials
import com.amazonaws.regions.Region
import com.amazonaws.regions.Regions
import com.amazonaws.services.s3.AmazonS3Client
import com.amazonaws.services.s3.model.Bucket
import com.amazonaws.services.s3.model.GetObjectRequest
import java.io.File
import java.util.Date
import java.util.function.Consumer


class AmazonS3Util {
    companion object {
        private const val TAG = "AmazonS3Util"
        private var accessKeyId = ""
        private var secretAccessKey = ""

        const val BUCKET_NFL_S3 = "nfl-s3"
        val BUCKET_NFL_S3_REGION = Regions.US_WEST_2

        fun init(context: Context) {
        }

        fun initializeS3Client(): AmazonS3Client? {
            val credentials = AnonymousAWSCredentials()
            return AmazonS3Client(credentials, Region.getRegion(BUCKET_NFL_S3_REGION))
        }

        suspend fun listBuckets(s3: AmazonS3Client) {
            try {
                val bucketList = s3.listBuckets()
                bucketList.forEach(Consumer { bucket: Bucket ->
                    Log.d(TAG, "Bucket Name: " + bucket.name)
                })
            } catch (e: Exception) {
                Log.e(TAG, "listBuckets error: $e")
            }
        }

        suspend fun createPresignedGetUrl(
            s3Client: AmazonS3Client,
            bucketName: String = BUCKET_NFL_S3,
            keyName: String,
            /**
             * Valid for 1 day.
             */
            expiration: Long = 24 * 60 * 60 * 1000
        ): String? {
            try {
                val time = expiration + System.currentTimeMillis()
                return s3Client.generatePresignedUrl(bucketName, keyName, Date(time)).let {
                    Log.d(TAG, "s3 url: $it")
                    it.toString()
                }
            } catch (e: Exception) {
                Log.e(TAG, "createPresignedGetUrl failed: $e")
            }
            return null
        }

        suspend fun downloadFile(
            s3: AmazonS3Client,
            bucketName: String,
            key: String,
            destinationPath: String
        ) {
            try {
                val getObjectRequest = GetObjectRequest(bucketName, key)
                s3.getObject(getObjectRequest, File(destinationPath))
                Log.d(TAG, "File downloaded to: $destinationPath")
            } catch (e: Exception) {
                Log.e(
                    TAG,
                    "downloadFile bucket name:$bucketName, key:$key, path:$destinationPath failed and msg: $e"
                )
            }

        }
    }
}