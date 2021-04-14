import React from 'react'
import PropTypes from 'prop-types'
import ClassNames from 'classnames'
import './index.scss'

const Button = (props) => {
	const style = ClassNames('button',
		`button--${props.type}`,
		`button--${props.size}`
	)

	return (
		<button onClick={(e) => props.onClick(e)} className={style} disabled={props.disabled}>
			{props.children}
		</button>
	)
}

Button.propTypes = {
	children: PropTypes.oneOfType([
		PropTypes.element,
		PropTypes.string
	]).isRequired,
	type: PropTypes.oneOf(['primary', 'secondary', 'read-more', 'footer', 'footer-leaked']),
	size: PropTypes.oneOf(['small', 'cta']),
	disabled: PropTypes.bool,
	onClick: PropTypes.func
}

Button.defaultProps = {
	type: 'primary'
}

export default Button
