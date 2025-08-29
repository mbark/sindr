import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
    title: string;
    Svg: React.ComponentType<React.ComponentProps<'svg'>>;
    description: ReactNode;
};

const FeatureList: FeatureItem[] = [
    {
        title: 'Build a CLI',
        Svg: require('@site/static/img/undraw_programming_65t2.svg').default,
        description: (
            <>
                No need to learn how to use the tool, you build a CLI with flags, arguments and auto completion.
            </>
        ),
    },
    {
        title: 'Configuration as code',
        Svg: require('@site/static/img/undraw_building-blocks_h5jb.svg').default,
        description: (
            <>
                No need to learn some arcane language or being restricted by a configuration language, configure your
                CLI by writing Starlark, a Python-subset designed for configuration.
            </>
        ),
    },
    {
        title: 'Batteries included',
        Svg: require('@site/static/img/undraw_outer-space_qey5.svg').default,
        description: (
            <>
                Contains everything you need for your commands like running shell commands, doing async operations,
                caching, templating, and more.
            </>
        ),
    },
];

function Feature({title, Svg, description}: FeatureItem) {
    return (
        <div className={clsx('col col--4')}>
            <div className="text--center">
                <Svg className={styles.featureSvg} role="img"/>
            </div>
            <div className="text--center padding-horiz--md">
                <Heading as="h3">{title}</Heading>
                <p>{description}</p>
            </div>
        </div>
    );
}

export default function HomepageFeatures(): ReactNode {
    return (
        <section className={styles.features}>
            <div className="container">
                <div className="row">
                    {FeatureList.map((props, idx) => (
                        <Feature key={idx} {...props} />
                    ))}
                </div>
            </div>
        </section>
    );
}
