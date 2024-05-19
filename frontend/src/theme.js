// src/theme.js
import { createTheme } from '@mui/material/styles';

const theme = createTheme({
  palette: {
    primary: {
    main: '#03045e', // Jasny różowy
},
    secondary: {
        main: '#00743f', // Jasny niebieski
    },
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
        //   color: 'white', // Kolor tekstu przycisku
          backgroundColor: "#",
          '&:hover': {
            backgroundColor: '#42a5f5', // Kolor tła przycisku po najechaniu
          },
        },
      },
    },
  },
});

export default theme;
