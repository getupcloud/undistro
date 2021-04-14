/* eslint-disable eqeqeq */
import React, { useState } from 'react'
import PropTypes from 'prop-types'
import { motion } from 'framer-motion'

import './index.scss'

const MarkdownAccordion = (props) => {
	const [show, setShow] = useState(false)

	return (
		<div className='markdown-navigation'>
			<div onClick={() => setShow(!show)} className='header'>
				<a href={props.id} className='title-anchor title-level1'>
					<span>{props.number}</span>
					{props.title}
				</a>
				{!props.subtitle.length == [] && <i onClick={() => setShow(!show)} className={show ? 'icon-arrow-up' : 'icon-arrow-down'} />}
			</div>

			{show && props.subtitle.map(sub => {
				return (
					<motion.div
						key={sub.id}
						className='subtitles'
						initial={{ y: -20 }}
						animate={{ y: 0 }}
						transition={{ ease: 'easeOut', duration: 0.2 }}
					>
						<a href={sub.id} className='title-anchor title-level2'>
							<span>{sub.number}</span>
							{sub.title}
						</a>
					</motion.div>
				)
			})}
		</div>
	)
}

export default MarkdownAccordion

MarkdownAccordion.propTypes = {
	subtitle: PropTypes.array,
	id: PropTypes.string,
	number: PropTypes.string,
	title: PropTypes.string
}

// title-anchor title-level2
