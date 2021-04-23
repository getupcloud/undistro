/* eslint-disable eqeqeq */
import React, { useState } from 'react'
import PropTypes from 'prop-types'
import { motion } from 'framer-motion'
import Classnames from 'classnames'
import './index.scss'

const MarkdownAccordion = (props) => {
	const [show, setShow] = useState(false)

	const toggle = (i) => {
		(show === i) ? setShow(null) : setShow(i)
	}

	return (
		<div className='markdown-navigation'>
			{props.navigation.map((elm, i) => {
				return (
					<>
						<div key={i} onClick={() => toggle(i)} className={Classnames('header', { active: show === i })}>
							<a href={elm.id} className='title-anchor title-level1'>
								<span>{elm.number}</span>
								{elm.title}
							</a>
							{!elm.subtitle.length == [] && <i onClick={() => toggle(i)} className={show === i ? 'icon-arrow-up' : 'icon-arrow-down'} />}
						</div>
						{show === i && elm.subtitle.map(sub => {
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
					</>
				)
			})}
		</div>
	)
}

export default MarkdownAccordion

MarkdownAccordion.propTypes = {
	navigation: PropTypes.func
}

// title-anchor title-level2
