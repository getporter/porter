resource "local_file" "foo" {
    content  = var.file_contents
    filename = "${path.module}/foo"
}