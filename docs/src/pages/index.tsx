import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import Heading from '@theme/Heading';

import styles from './index.module.css';
import HomepageCode from '../components/HomepageCode';

function HomepageHeader() {
    const {siteConfig} = useDocusaurusContext();
    return (
        <header className={clsx(styles.heroBanner)}>
            <div className="container">
                <Heading as="h1" className="hero__title">
                    {siteConfig.title}
                </Heading>
                <p className="hero__subtitle">{siteConfig.tagline}</p>
                <div className={styles.buttons}>
                    <Link
                        className="button button--secondary button--lg"
                        to="/docs/installation">
                        🔨 Get started
                    </Link>
                    <Link
                        className="button button--outline button--secondary button--lg"
                        to="/docs/getting-started"
                        style={{marginLeft: '1rem'}}>
                        Learn more
                    </Link>
                </div>
            </div>
        </header>
    );
}

export default function Home(): ReactNode {
    const {siteConfig} = useDocusaurusContext();
    return (
        <Layout
            title={`${siteConfig.title} - Project-specific commands as a CLI`}
            description="Project-specific commands as a CLI.">
            <HomepageHeader/>
            <main>
                <HomepageCode/>
                <HomepageFeatures/>
            </main>
        </Layout>
    );
}
