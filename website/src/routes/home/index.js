import React from 'react'
import CallToAction from 'Components/molecules/callToAction'
import KeyBenefits from 'Components/atoms/keyBenefits'
import Highlight from 'Components/atoms/highlight'
import Roadmap from 'Components/atoms/roadmap'
import FAQ from 'Components/atoms/faq'
import Button from 'Components/atoms/button'
import { useHistory } from 'react-router'

import './index.scss'

export default function HomePageRoute () {
	const history = useHistory()

	return (
		<div className='main-container'>
			<CallToAction />
			<KeyBenefits />
			<Highlight />
			<Roadmap />
			<FAQ />
			<div className='buttons-cta'>
				<Button onClick={() => window.open('https://github.com/getupio-undistro/undistro/releases')} type='footer'>DOWNLOAD</Button>
				<Button onClick={() => history.push('/docs#3---installing-undistro')} type='footer-leaked'>GET STARTED</Button>
			</div>
		</div>
	)
}
