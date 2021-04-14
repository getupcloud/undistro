import React, { useState, useEffect } from 'react'
import PropTypes from 'prop-types'
import Classnames from 'classnames'
import { isMobile } from 'react-device-detect'
import './index.scss'

const LIMIT = isMobile ? 6 : 8

const Pagination = (props) => {
	const [pages, setPages] = useState([])

	function getPagesData () {
		const data = []

		if (props.total > LIMIT) {
			if (props.current < LIMIT - 3) {
				// in left
				for (let i = 1; i <= LIMIT - 2; i++) {
					data.push({ page: i, class: [i === props.current ? 'active' : ''] })
				}

				data.push({ page: '...', class: 'disabled' })
				data.push({ page: props.total })

				return data
			} else if (props.current > props.total - 3) {
				// in right
				data.push({ page: 1 })
				data.push({ page: '...', class: 'disabled' })

				for (let i = props.total - LIMIT + 3; i <= props.total; i++) {
					data.push({ page: i, class: [i === props.current ? 'active' : ''] })
				}

				return data
			}

			if (props.current >= 3 && props.current <= props.total - 3) {
				// in middle
				// data = []
				data.push({ page: 1 })
				data.push({ page: '...', class: 'disabled' })

				const space = Math.floor((LIMIT - 4) / 2) || 1

				for (let i = props.current - space; i <= props.current + space; i++) {
					data.push({ page: i, class: [i === props.current ? 'active' : ''] })
				}

				data.push({ page: '...', class: 'disabled' })
				data.push({ page: props.total })

				return data
			}
		} else {
			for (let i = 1; i <= props.total; i++) {
				data.push({ page: i, class: [i === props.current ? 'active' : ''] })
			}

			return data
		}
	}

	useEffect(() => {
		setPages(getPagesData())
	}, [props.current, props.total])

	const handleAction = (elm) => {
		if (elm.page !== '...') {
			props.onChange(elm.page)
		}
	}

	const handleLeft = () => {
		if (props.current > 1) {
			props.onChange(props.current - 1)
		}
	}

	const handleRight = () => {
		if (props.current < props.total) {
			props.onChange(props.current + 1)
		}
	}

	if (props.total < 2) return null

	return (
		<div className='pagination'>
			<i className='icon-arrow-left' onClick={handleLeft} />
			{pages.map((elm, i) =>
				<div onClick={() => handleAction(elm)} key={i} className={Classnames('item', elm.class)}>
					{elm.page}
				</div>
			)}
			<i className='icon-arrow-right' onClick={handleRight} />
		</div>
	)
}

Pagination.propTypes = {
	total: PropTypes.number,
	current: PropTypes.number,
	onChange: PropTypes.func
}

Pagination.defaultProps = {
	total: 1,
	current: 1
}

export default Pagination
