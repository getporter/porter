---
title: When to use credentials vs parameters
description: When writing a bundle in Porter, when should you use credentials instead of parameters?
weight: 1
---

Porter has two types of bundle inputs: credentials and parameters.
They both appear similar from the bundle authoring experience. 
They are injected into a bundle as either environment variables or files.
They are mapped from parameter and credential sets that have similar capabilities for sourcing secrets from external secret stores or from the host environment.
Both can hold sensitive data... and yet internally they are treated very differently by Porter.

Here are the general rules for when to use credentials or parameters:

* **Use credentials for information that is associated to a person, such as your username or password.**
Porter never remembers credentials and always requires that they are passed in again when you run the bundle.
* **Use parameters for information that can be shared between the people running a bundle.**
This includes sensitive data such as your application's database connection string, configuration values, or ca certificates.
Porter remembers previously used parameters and can reuse them when the bundle is run again.

For more context on this recommendation, we need to look at the CNAB specification:

* Credentials are identifying pieces of information that are associated with the person running the bundle.
If someone upgrades an installation and provides their credentials to an environment or cloud provider, those credentials are specific to that user.
If a different person upgrades an installation later that week, they would use their own credentials.
* Parameters are inputs to the bundle that do not vary based on the person who is running the bundle.
Parameters may provide sensitive data, such as a mysql connection string, but regardless of who is running the bundle the same connection string is passed to the bundle.

The CNAB specification requires that Porter never persist credentials.
They should only be injected just-in-time into the bundle, and if the bundle needs to be re-run, the credentials must be provided again. 
Parameters are required by the CNAB specification to be persisted so that they can be reused when the bundle is re-run.
Porter remembers the parameters passed when you originally installed the bundle, and in Porter v1 also reuses those values when you upgrade the bundle.

The guidelines above are only recommendations.
You may choose to only use credentials for sensitive data, and take advantage of the Porter's hands-off treatment of credentials so that you are assured that Porter will never persist that data.
It really depends on your security concerns and how much the changes in the user experience are relevant to your scenario.

## Warning
⚠️ In versions of Porter before [v1.0.0-alpha.20], Porter persists ALL parameters in its database, sensitive or otherwise.
In v1.0.0-alpha.20 and higher, Porter does not save sensitive data in its database and stores them in an external secret store instead.
Read [Upgrade Your Plugins to securely store sensitive data](/blog/persist-sensitive-data-safely/) to learn more. 

[v1.0.0-alpha.20]: https://github.com/getporter/porter/releases/tag/v1.0.0-alpha.20

## See Also

* [Credentials Overview](/credentials/)
* [Parameters Overview](/parameters/)
