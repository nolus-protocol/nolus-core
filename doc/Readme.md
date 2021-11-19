# Documentation
The documentation employs the [C4 Model](https://c4model.com/) technique to model and visualize the system architecture from its static and dynamic perspective.

This repo contains:
- the main C4 workspace,
- additional documentation in form of Markdown files, and
- Architectural Decision Records (ADRs) expressed in [MADR](https://adr.github.io/madr/) format.

It has hierarchical organization mirroring the static system structure. For example, the documentation of container "Smart Contracts" is located in `smart-contracts` directory.

## Authoring
The easiest way to create and edit the model and views is using Structurizr DSL. The [Web console](https://structurizr.com/dsl) provides an editor and a convenient side view. Other option is to use any text editor of preference. In either way, the [DSL reference](https://github.com/structurizr/dsl/blob/master/docs/language-reference.md) is helpful.

## Rendering
Locally, the views are edited by using Structurizr Lite. It is a Java Tomcat Web Application allowing one to edit the layout of the views and export to PNG and SVG. The changes are saved as a workspace in `workspace.json` file next to the source `workspace.dsl`.

    docker pull structurizr/lite
    docker run -it --rm -p 8080:8080 -v <workspace-dir>:/usr/local/structurizr structurizr/lite

## References
[Technology Radar's Diagrams-as-Code](https://www.thoughtworks.com/radar/techniques/diagrams-as-code)
[C4 Model](https://c4model.com/)
[Structurizr](https://structurizr.org/)
[Getting Started](https://dev.to/simonbrown/getting-started-with-structurizr-lite-27d0)