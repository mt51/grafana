"""
This module returns all the pipelines used in the event of documentation changes along with supporting functions.
"""

load(
    "scripts/drone/steps/lib.star",
    "build_docs_website_step",
    "build_image",
    "codespell_step",
    "identify_runner_step",
    "yarn_install_step",
)
load(
    "scripts/drone/utils/utils.star",
    "pipeline",
)

docs_paths = {
    "include": [
        "*.md",
        "docs/**",
        "packages/**/*.md",
        "latest.json",
    ],
}

def docs_pipelines(ver_mode, trigger):
    environment = {"EDITION": "oss"}
    steps = [
        identify_runner_step(),
        yarn_install_step(),
        codespell_step(),
        build_docs_website_step(),
    ]

    return pipeline(
        name = "{}-docs".format(ver_mode),
        edition = "oss",
        trigger = trigger,
        services = [],
        steps = steps,
        environment = environment,
    )

def trigger_docs_main():
    return {
        "branch": "main",
        "event": [
            "push",
        ],
        "paths": docs_paths,
    }

def trigger_docs_pr():
    return {
        "event": [
            "pull_request",
        ],
        "paths": docs_paths,
    }
