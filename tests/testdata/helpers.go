package testdata

const (
	// MyBuns is our standard test bundle with a bunch of features in it.
	MyBuns = "mybuns"

	// MyBunsRef is the full reference to the mybuns test bundle.
	MyBunsRef = "localhost:5000/mybuns:v0.1.2"

	// MyDb is the test bundle that is a dependency of mybuns.
	MyDb = "mydb"

	// MyDbRef is the full reference to the mydb test bundle.
	MyDbRef = "localhost:5000/mydb:v0.1.0"

	// MyBunsWithImgReference is the test bundle that contains image reference.
	MyBunsWithImgReference = "mybun-with-img-reference"

	// MyBunsWithImgReference is the full reference to the test bundle that contains image reference.
	MyBunsWithImgReferenceRef = "localhost:5000/mybun-with-img-reference:v0.1.0"
)
