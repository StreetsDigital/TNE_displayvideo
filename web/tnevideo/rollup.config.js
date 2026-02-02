import terser from '@rollup/plugin-terser';

const isMinify = process.env.MINIFY === 'true';

export default {
  input: 'src/tnevideo.js',
  output: {
    file: isMinify ? 'dist/tnevideo.min.js' : 'dist/tnevideo.js',
    format: 'iife',
    name: 'TNEVideo',
    banner: '/* TNEVideo v1.0.0 | (c) The Nexus Engine */'
  },
  plugins: isMinify ? [terser()] : []
};
