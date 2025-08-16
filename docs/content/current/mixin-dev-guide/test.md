---
title: Test a Mixin
description: How to test a Porter mixin
aliases:
- /mixin-dev-guide/testing/
---

We are working on filling this doc out more, until then this is more of a FAQ than a proscriptive guide. If you have
a tip, please submit a PR and help us fill this out!

## How do I unit test that my mixin is executing the right commands?

Here is a [full working example][example] of a unit test that validates the commands executed by a mixin.

Make sure that your package has a `TestMain` that calls `github.com/getporter/porter/pkg/test.TestMainWithMockedCommandHandlers`

```go
import "get.porter.sh/porter/pkg/test"

func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}
```

Then you instantiate your mixin in test mode (the skeletor template generates this method for you):

```go
m := NewTestMixin(t)
```

This sets up your test binary to assert that expected command(s) were called. You tell it what command to expect with

```go
m.Setenv(test.ExpectedCommandEnv, "helm install")
```

If your mixin action executes multiple commands, separate them with a newline `\n` like so

```go
m.Setenv(test.ExpectedCommandEnv, "helm install\nhelm upgrade")
```

Now execute your mixin action:

```go
err = m.Execute()
```

Instead of os calls to the real commands, the test mixin mode calls back into your test binary. The `TestMain` handles
asserting that the expected commands were made and fails the test if they weren't.

[example]: https://github.com/getporter/gcloud-mixin/blob/v0.2.1-beta.1/pkg/gcloud/execute_test.go
