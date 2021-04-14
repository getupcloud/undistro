import React from 'react'
import PropTypes from 'prop-types'

import './index.scss'

const HorizontalCard = (props) => {
	return (
		<div onClick={props.onClick} className='horizontal-card-container'>
			<div className='left-text'>
				<span>{props.title}</span>
				<p>{props.title2}</p>
			</div>

			<div className='right-text'>
				{props.description}
			</div>
		</div>
	)
}

HorizontalCard.propTypes = {
	title: PropTypes.string,
	title2: PropTypes.string,
	description: PropTypes.string,
	onClick: PropTypes.func
}

export default HorizontalCard
