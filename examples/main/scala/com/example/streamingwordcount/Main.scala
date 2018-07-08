package com.example.streamingwordcount

import org.apache.spark.sql.SparkSession

/**
  *
  */
object Main {


	def main(args: Array[String]): Unit = {

		val spark: SparkSession = SparkSession.builder
			.master(System.getProperty("spark.master", "local[*]"))
			.appName(getClass.getName)
			.getOrCreate

		import spark.implicits._

		val lines = spark.readStream
			.format("socket")
			.option("host", args.headOption.getOrElse("localhost"))
			.option("port", 9999)
			.load()

		// Split the lines into words
		val words = lines.as[String]
			.flatMap(_.split(" "))
	    //	.repartition(5)
		//.map("some-prefix-" + _)

		// Generate running word count
		val wordCounts = words.groupBy("value").count()

		// Start running the query that prints the running counts to the console
		val query = wordCounts.writeStream
			.outputMode("complete")
			.format("console")
			.start()

		query.awaitTermination()
	}

}
