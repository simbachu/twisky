The raw file is the Inkscape working document (multipage). Label each page
`icon-<name>`; drawable groups on a layer are matched by the same
`inkscape:label`, or by translate origin falling inside the page.

Pack into `static/web/icons/icons.svg` with:

```bash
go generate ./static
```

CI and Docker run the same generate step. Enable the repo git hooks so packing
is checked when the raw file is staged:

```bash
git config core.hooksPath .githooks
```
