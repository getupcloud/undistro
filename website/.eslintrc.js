/* eslint-disable semi */
module.exports = {
	parser: 'babel-eslint',
	extends: ['standard', 'standard-react', 'plugin:react/recommended'],
	rules: {
		semi: [2, 'never'],
		quotes: ['warn', 'single'],
		indent: [2, 'tab'],
		'no-tabs': 0
	}
};
