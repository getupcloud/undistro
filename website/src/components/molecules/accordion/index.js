import React, { useState } from 'react'
import PropTypes from 'prop-types'
import { motion } from 'framer-motion'

import './index.scss'

const Accordion = (props) => {
	const [show, setShow] = useState(false)

	return (
		<div className='accordion-container'>
			<div onClick={() => setShow(!show)} className='header'>
				<div className='question'>
					<p>{props.question}</p>
				</div>
				<i className={show ? 'icon-arrow-up' : 'icon-arrow-down'} />
			</div>

			{show &&
				<motion.div
					initial={{ y: -20 }}
					animate={{ y: 0 }}
					transition={{ ease: 'easeOut', duration: 0.2 }}
					className='answer'
				>
					<p>{props.answer}</p>
				</motion.div>}
		</div>
	)
}

Accordion.propTypes = {
	question: PropTypes.string,
	answer: PropTypes.string
}

export default Accordion
