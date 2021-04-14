import React from 'react'
import Getup from 'Assets/images/getup.svg'

import './index.scss'

const Footer = () => {
	const redirect = (link) => {
		switch (link) {
		case 'twitter':
			return window.open('https://twitter.com/undistro')
		case 'git':
			return window.open('https://github.com/getupio-undistro/undistro')
		case 'insta':
			return window.open('https://www.instagram.com/undistro.io/')
		default:
			window.open('http://getup.io/en')
		}
	}

	return (
		<div className='footer'>
			<img src={Getup} onClick={redirect} />

			<div className='socials'>
				<i onClick={() => redirect('insta')} className='icon-insta' />
				<i onClick={() => redirect('twitter')} className='icon-twitter' />
				<i className='icon-slack' />
				<i onClick={() => redirect('git')} className='icon-github' />
			</div>
		</div>
	)
}

export default Footer
