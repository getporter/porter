# Does Porter Replace Duffle?

**TL;DR: NOPE 😁**

## What is Duffle?

[Duffle](https://github.com/deislabs/duffle) is a command-line tool that can be used to build and install [Cloud Native Application Bundles](https://cnab.io) (CNABs). Using Duffle, bundle authors can build, test and distribute installers for their application bundles. The CNAB specification provides high-level guidance for the structure of a bundle, but gives the author a great deal of flexibility. Duffle is a reference implementaiton of the spec and isn't intended to be opinionated. So it also has a great deal of flexibility, at the cost of requiring more knowledge of the CNAB spec when authoring bundles. 

## What is Porter?

Porter is also a command-line tool that can be used to build bundles. It uses a declarative manifest and a special runtime paradigm to make it easier to quickly author bundles, and compose bundles from other porter authored bundles.

## Why Porter?

If Duffle and Porter can both build bundles, why would you use Porter? 

As mentioned above, the CNAB specification defines the general runtime structure of an application bundle. This enables tools like Duffle to support both building and installing bundles. Unfortuanately, with great flexibility comes great complexity. 

Porter on the other hand, provides a declarative authoring experience that encourages the user to adhere to its opionated bundle design. Porter introduces a structured manifest that allows bundle authors to declare dependencies on other bundles, explicitly declare the capabilities that a bundle will use and how parameters, credentials and outputs are passed to individual steps within a bundle. This allows bundle authors to create reusable bundles without requiring extensive knowledge of the CNAB spec.

Porter introduces a command-line tool, along with buildtime and runtime components, called mixins. Mixins give CNAB authors smart componenents that understand how to adapt existing systems, such as Helm, Terraform or Azure, into CNAB actions. After creating a Porter manifest, the command-line tool creates an invocation image and the require CNAB bundle.json. Porter utilizes each mixin to determine the contents of the invocation image. The Porter build command also adds the porter manifest and the porter runtime functionality into the invocation image.  

After a Porter authored bundle is built, it can be run by Duffle or any other CNAB compliant tool. When the bundle is installed, the Porter runtime uses the bundle's porter manifest to determine how to invoke the runtime functionality of each mixin, including how to pass bundle parameters and credentials and how to wire their ouputs to other steps in the bundle. This capability also allows bundle authors to declare dependencies on other Porter bundles and pass the output from one to another.

Porter is **not** intended to replace Duffle, but instead exists to provide an improved bundle authoring experience.



