import React, { useState } from 'react'
import { useHistory } from 'react-router'
import { motion } from 'framer-motion'

import './index.scss'

const HamburgerMenu = () => {
	const [show, setShow] = useState(false)
	const history = useHistory()

	const redirect = (url) => {
		history.push(`${url}`)
		setShow(false)
	}
	return (
		<div className='hamburger-menu'>
			<i onClick={() => setShow(!show)} className='icon-hamburger' />

			{show &&
        <motion.ul initial={{ x: -100 }} animate={{ x: 0 }} transition={{ ease: 'easeOut', duration: 0.2 }}>
        	<li className='options' onClick={() => redirect('/')}>Home</li>
        	<li className='options' onClick={() => redirect('/docs')}>Docs</li>
        	{/* <li>Tutorials</li> */}
        	{/* <li onClick={() => window.location.replace('https://blog.getupcloud.com/')}>Blog</li> */}
        	<li className='options' onClick={() => redirect('/community')}>Community</li>
        	<li className='options' onClick={() => redirect('/faq')}>FAQ</li>
        </motion.ul>}
		</div>
	)
}

export default HamburgerMenu
