# Does Porter Replace Duffle?

**TL;DR: NOPE 😁**

## What is Duffle?

Duffle is a command line tool that can be used to build and install CNABs. Using Duffle, bundle authors can build, test and distribute CNABs for installing their cloud native applications. The CNAB specification provides high level specification for the structure of a CNAB, but gives the bundle author a great deal of flexibility. Duffle also provides the author this flexibility.  

## What is Porter?

Porter is also a command line tool that can be used to build CNABs, but enforces a declarative build and runtime paradigm. 

## Why Porter?

If [Duffle](https://github.com/deislabs/duffle) and Porter can both build CNABs, why was Porter created? 

As mentioned above, the CNAB specification defines the general structure of an application bundle. This enables tools like Duffle to support both building and installing bundles. With great flexibility can come great complexity, however. Building CNABs can be a difficult undertaking. 

Porter, on the other hand, provides a declarative authoring experience. Porter introduces a structured manifest that allows bundle authors to declare dependencies on other bundles, explicitly declare the capabilities that a bundle will use and how parameters, credentials and outputs are passed to individual steps within a bundle. This allows bundle authors to create reusable bundles that translate CNAB actions into Helm, Terraform, Azure or other systems. 

Porter introduces a command line tool, along with build time and run time components, called mixins. Mixins enable CNAB authors to declaratively define a bundle by specifying the functionality that they need without needing to write shell scripts or add things into the invocation image. After declaring the functionality of a bundle, the Porter command line tool is used to create an invocation image and the bundle.json. Porter utilizes each mixin to determine the contents of the invocation image. The Porter build also adds the porter manifest and the porter runtime functionality to the invocation image.  

After a Porter bundle has been built, however, it can be run by Duffle or any other CNAB compliant tool. When a Porter bundle is run, the Porter runtime uses the same manifest that was used at build time to determine how to invoke the runtime functionality of each mixin, including how to pass bundle parameters and credentials and what to do with their outputs. This capability also allows bundle authors to declare dependencies on other Porter bundles and pass the output from one to another.

Porter is **not** intended to replace Duffle, but instead exists to provide a different bundle authoring experience.



