import React from 'react'
import Button from 'Components/atoms/button'
import { useHistory } from 'react-router'

import './index.scss'

const CallToAction = () => {
	const history = useHistory()

	return (
		<div className='cta-background'>
			<div className='cta-container'>
				<div className='content'>
					<h1>CENTRALIZED AND STANDARDIZED KUBERNETES OPERATIONS</h1>
					<div className='buttons'>
						<Button onClick={() => window.open('https://github.com/getupio-undistro/undistro/releases')} size='cta'>DOWNLOAD</Button>
						<Button onClick={() => history.push('/docs#3---installing-undistro')} size='cta' type='secondary'>GET STARTED</Button>
					</div>
				</div>
			</div>
		</div>
	)
}

export default CallToAction
