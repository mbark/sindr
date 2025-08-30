import type {ReactNode} from 'react';
import SyntaxHighlighter from 'react-syntax-highlighter';
import {stackoverflowDark, stackoverflowLight} from 'react-syntax-highlighter/dist/esm/styles/hljs';
import clsx from "clsx";
import style from "./styles.module.css";

export default function HomepageFeatures(): ReactNode {
    return (
        <section>
            <div className="container">
                <div className={clsx(style.code, "row")}>
                    <CodeComponent code={sindrCode} language="python"/>
                </div>
                <div className={clsx(style.code, "row")}>
                    <CodeComponent code={makefileCode} language="makefile"/>
                </div>
            </div>
        </section>
    );
}

const CodeComponent: React.FC<{code: string, language: string}> =  ({code, language}) =>  {
    return <>
        <SyntaxHighlighter language={language} style={stackoverflowLight} data-theme-is={"light"}>
            {code}
        </SyntaxHighlighter>
        <SyntaxHighlighter language={language} style={stackoverflowDark} data-theme-is={"dark"}>
            {code}
        </SyntaxHighlighter>
    </>
}

const sindrCode = `
cli(
    name = "sindr",
    usage = "✨🔨 Project-specific commands as a CLI. "
)

def test(ctx):
    shell('go test {{.flags}} {{.args}} ./...',
        flags='-short' if ctx.short else '')

command(
    name = "test",
    usage = "run go test",
    action = test,
    args = ['args'],
    flags = [
        {
            "name": "short",
            "type": "bool",
            "default": True,
            "usage": "Use the -short flag when running the tests"
        },
    ],
)`

const makefileCode = `
SHORT ?= true
ARGS ?=

test:
	@if [ "$(SHORT)" = "true" ]; then \\
		go test -short $(ARGS) ./...; \\
	else \\
		go test $(ARGS) ./...; \\
	fi
.PHONY: test

help:
	@echo "Available targets:"
	@echo "  test    - run go test"
	@echo ""
	@echo "Variables:"
	@echo "  SHORT   - Use the -short flag when running the tests (default: true)"
	@echo "  ARGS    - Additional arguments to pass to go test"
	@echo ""
	@echo "Examples:"
	@echo "  make test"
	@echo "  make test SHORT=false"
	@echo "  make test ARGS='-v -race'"
`
