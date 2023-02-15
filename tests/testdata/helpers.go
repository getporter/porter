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

	// MyEnv is the root test bundle that exercises dependencies.
	MyEnv = "myenv"

	// MyEnvRef is the full reference to the myenv test bundle.
	MyEnvRef = "localhost:5000/myenv:v0.1.0"

	// MyInfra is the root test bundle that exercises dependencies.
	MyInfra = "myinfra"

	// MyInfraRef is the full reference to the myinfra test bundle.
	MyInfraRef = "localhost:5000/myinfra:v0.1.0"

	// MyApp is the root test bundle that exercises dependencies.
	MyApp = "myapp"

	// MyAppRef is the full reference to the myapp test bundle.
	MyAppRef = "localhost:5000/myapp:v1.2.3"
)
