package spark.internal;

import java.io.*;
import java.util.Comparator;
import java.util.Iterator;
import java.util.jar.*;
import java.util.zip.ZipEntry;
import java.util.zip.ZipOutputStream;

/**
 *
 */
public class RepackageJar {

	public static void main(String[] args) throws IOException {
		repackage(args[0], args[1]);
		System.exit(0);
	}

	static void repackage(String srcPath, String targetPath) throws IOException {
		// Set up source & target jars
		JarFile src = new JarFile(srcPath);
		JarOutputStream target = new JarOutputStream(new BufferedOutputStream(new FileOutputStream(targetPath)));
		target.setMethod(ZipOutputStream.STORED);

		// Get a sorted list of source jar's entries
		Iterator<JarEntry> entries = src.stream()
				.sorted(Comparator.comparing(ZipEntry::getName))
				.iterator();

		// Copy over entries, squashing any metadata
		byte[] dataBuffer = new byte[2048];
		int dataLen;
		while (entries.hasNext()) {
			JarEntry e = entries.next();
			InputStream is = src.getInputStream(e);
			e.setMethod(ZipEntry.STORED);
			e.setCompressedSize(e.getSize());
			e.setTime(0);
			target.putNextEntry(e);
			while ((dataLen = is.read(dataBuffer)) > 0) {
				target.write(dataBuffer, 0, dataLen);
			}
		}
		target.close();
	}

}
