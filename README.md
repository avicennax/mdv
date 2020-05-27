# mdv

A janky little CLI markdown viewer\*

----

Opens your default browser to view a Pandoc-generated HTML of
your markdown. You can bring your own CSS if you want.

As a little bonus it also caches HTML renders of the markdown source by mapping checksums of the markdown files to paths of the rendered HTML files - thereby only calling `pandoc` if the source has changed.

\*Technically the CLI tool just calls Pandoc and opens the generated file
with your default application for opening HTML files.
