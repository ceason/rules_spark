package com.example.periodicpi

import org.apache.spark.sql.SparkSession

/**
  *
  */
object Main {


	def main(args: Array[String]): Unit = {

		val numSamples = sys.env("NUM_SAMPLES").toInt

		val spark: SparkSession = SparkSession.builder
			.master(System.getProperty("spark.master", "local[*]"))
			.appName(getClass.getName)
			.getOrCreate

		val count = spark.sparkContext
			.parallelize(1 to numSamples)
			.repartition(5)
			.filter { _ =>
				val x = math.random
				val y = math.random
				x * x + y * y < 1
			}.count()
		println(s"Pi is roughly ${4.0 * count / numSamples}")
		Thread.sleep(30000)
	}

}
