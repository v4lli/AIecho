import {useMediaQuery} from 'react-responsive';

const Desktop = ({children}) => {
    const isDesktop = useMediaQuery({minWidth: 1224});
    return isDesktop ? children : null;
};

const MobilePortrait = ({children}) => {
    const isMobile = useMediaQuery({maxWidth: 1223});
    const isPortrait = useMediaQuery({orientation: "portrait"});
    return isMobile && isPortrait ? children : null;
};

const MobileLandscape = ({children}) => {
    const isMobile = useMediaQuery({maxWidth: 1223});
    const isLandscape = useMediaQuery({orientation: "landscape"});
    return isMobile && isLandscape ? children : null;
}

export {Desktop, MobileLandscape, MobilePortrait};
