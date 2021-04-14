import React from 'react'
import PropTypes from 'prop-types'

import './index.scss'

const Search = (props) => {
	return (
		<div className='search'>
			<input
				type={props.type}
				placeholder={props.placeholder}
				onChange={e => props.onChange(e.target.value)}
			/>
			<i className='icon-search' onClick={props.onClick} />
		</div>
	)
}

Search.propTypes = {
	placeholder: PropTypes.string,
	type: PropTypes.string,
	onChange: PropTypes.func,
	onClick: PropTypes.func
}

Search.defaultProps = {
	type: 'text'
}

export default Search
