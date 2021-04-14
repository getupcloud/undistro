import React from 'react'
import PropTypes from 'prop-types'
import { useHistory } from 'react-router-dom'
import Hamburger from 'Components/atoms/hamburgerMenu'

import './index.scss'

const Navbar = (props) => {
	const history = useHistory()

	const redirect = (url) => {
		history.replace(`${url}`)
	}

	return (
		<div className='navbar-container'>
			<Hamburger />
			<div className='img-container'>
				<img onClick={props.onClick} src={props.img} />
			</div>
			<ul className='list'>
				<li className='options' onClick={() => redirect('/')}>Home</li>
				<li className='options' onClick={() => redirect('/docs')}>Docs</li>
				{/* <li>Tutorials</li> */}
				{/* <li onClick={() => window.location.replace('https://blog.getupcloud.com/')}>Blog</li> */}
				<li className='options' onClick={() => redirect('/community')}>Community</li>
				<li className='options' onClick={() => redirect('/faq')}>FAQ</li>
			</ul>
		</div>
	)
}

Navbar.propTypes = {
	img: PropTypes.string,
	onClick: PropTypes.func
}

export default Navbar
